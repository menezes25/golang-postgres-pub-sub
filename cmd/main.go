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

	if err := run(logger); err != nil {
		logger.Fatal(err.Error())
	}
	logger.Info("num of go routines hangin ", runtime.NumGoroutine())
}

func run(l *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	postgresCli, err := postgres.NewProduction(db_host, db_name, db_user, db_port, db_pass, l)
	if err != nil {
		return err
	}

	notificationChan, err := postgresCli.ListenToEvents(ctx, "event")
	if err != nil {
		return err
	}

	responseChan := gopostgrespubsub.HandleEventEvents(notificationChan, l)

	wsManager := websocket.New(responseChan)

	mux := http.NewServeMux()

	mux.Handle("/api/event", rest.MakePostEventHandler(postgresCli))
	mux.Handle("/ws/echo/event", wsManager.MakeListenToEventsHandler())

	srv := rest.NewServer(fmt.Sprintf("0.0.0.0:%s", os.Getenv("REST_ADDRESS")), mux)

	errchan := make(chan error)
	go rest.StartServer(srv, l.With("vision_server", "rest"), errchan)
	l.Info("running rest server")

	doneChan := make(chan os.Signal, 1)
	signal.Notify(doneChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case oscall := <-doneChan:
		l.Infow("shuting servers down with", "sys_signal", oscall.String())
		rest.ShutdownServer(ctx, srv, l.With("server", "rest"))
	case err := <-errchan:
		l.Info("rest server crashed with ", err.Error())
	}

	//* finaliza execução das subrotinas
	cancel()

	//* espera finalização das subrotinas
	select {
	case <-time.After(20 * time.Second):
		return errors.New("shutdown timedout")
	case <-notificationChan:
		return nil
	}
}
