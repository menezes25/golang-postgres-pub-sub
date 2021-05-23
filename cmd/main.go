package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	gopostgrespubsub "postgres_pub_sub"
	"postgres_pub_sub/postgres"
	"postgres_pub_sub/trasnport/rest"
	"postgres_pub_sub/trasnport/websocket"
	"runtime"
	"syscall"
	"time"

	"go.uber.org/zap"
)

const (
	db_host = "localhost"
	db_name = "event"
	db_user = "tester"
	db_port = "5432"
	db_pass = "tester"
)

func main() {
	// https://github.com/uber-go/zap/issues/584
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.OutputPaths = []string{"stdout"}
	zapLogger, err := loggerConfig.Build()
	if err != nil {
		log.Fatal("failed to get logger with", err.Error())
	}
	logger := zapLogger.Sugar()

	err = run(logger)
	if err != nil {
		logger.Fatal(err.Error())
	}

	fmt.Printf("Number of hanging goroutines: %d", runtime.NumGoroutine()-1)
}

func run(l *zap.SugaredLogger) error {
	// ctx, cancel := context.WithCancel(context.Background())
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	postgresCli, err := postgres.NewProduction(db_host, db_name, db_user, db_port, db_pass, l)
	if err != nil {
		return err
	}
	postgresEventBus := gopostgrespubsub.NewEventBus()
	postgresCli.WithEventBus(postgresEventBus)

	//* jeito novo de fazer
	err = postgresCli.StartListeningToNotifications(ctx, []string{"event", "documents"})
	if err != nil {
		return err
	}

	dataEventChan := make(chan gopostgrespubsub.DataEvent, 1)
	postgresEventBus.Subscribe("event", dataEventChan)

	newResponseChan := gopostgrespubsub.HandleEventData(ctx, dataEventChan, l)

	newWsManager := websocket.New(newResponseChan)

	mux := http.NewServeMux()

	mux.Handle("/api/event", rest.MakePostEventHandler(postgresCli))
	mux.Handle("/ws/new/event", newWsManager.MakeListenToEventsHandler())

	srv := rest.NewServer(fmt.Sprintf("0.0.0.0:%s", os.Getenv("REST_ADDRESS")), mux)

	errchan := make(chan error, 1)
	go rest.StartServer(srv, l.With("vision_server", "rest"), errchan)
	l.Info("running rest server")

	select {
	case <-ctx.Done():
		l.Infow("shuting servers down with", "error", ctx.Err().Error())
		rest.ShutdownServer(context.Background(), srv, l.With("server", "rest"))
	case err := <-errchan:
		l.Info("rest server crashed with ", err.Error())
	}

	//* finaliza execução das subrotinas
	cancel()

	//* espera finalização das subrotinas
	select {
	case <-time.After(20 * time.Second):
		return errors.New("shutdown timedout")
	case <-newResponseChan:
		return nil
	}
}
