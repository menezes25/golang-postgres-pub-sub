package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	gopostgrespubsub "postgres_pub_sub"
	"postgres_pub_sub/postgres"
	transporthttp "postgres_pub_sub/transport/http"
	"runtime"
	"syscall"
	"time"

	"go.uber.org/zap"
)

const (
	db_host = "localhost"
	db_name = "event"
	db_user = "tester"
	db_port = "5432" // docker-compose 15432
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
	fmt.Printf("Number of hanging goroutines: %d", runtime.NumGoroutine()-1)
	if err != nil {
		logger.Fatal(err.Error())
	}
}

func run(l *zap.SugaredLogger) error {
	l.Infof("Starting PUB/SUB, PID: %d", os.Getpid())
	// ctx, cancel := context.WithCancel(context.Background())
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	postgresCli, err := postgres.NewProduction(db_host, db_name, db_user, db_port, db_pass, l)
	if err != nil {
		return err
	}

	const topicEvent = "event"
	const topicBoleto = "boleto"
	postgresEventBus := gopostgrespubsub.NewEventBus([]string{topicEvent, topicBoleto})
	postgresCli.WithEventBus(postgresEventBus)

	err = postgresCli.StartListeningToNotifications(ctx)
	if err != nil {
		return err
	}

	dataEventChan := postgresEventBus.Subscribe(topicEvent)
	dataBoletoChan := postgresEventBus.Subscribe(topicBoleto)
	fanInEvent := gopostgrespubsub.Merge(
		gopostgrespubsub.HandleData(ctx, dataEventChan, l, topicEvent, gopostgrespubsub.HandlePostgresDataEvent),
		gopostgrespubsub.HandleData(ctx, dataBoletoChan, l, topicBoleto, gopostgrespubsub.HandlePostgresDataBoleto),
	)

	transHttpRouter, err := transporthttp.NewTransportHttp(ctx, postgresCli, fanInEvent)
	if err != nil {
		return err
	}

	srv := transporthttp.NewServer(fmt.Sprintf("0.0.0.0:%s", os.Getenv("REST_ADDRESS")), transHttpRouter)

	errchan := make(chan error, 1)
	go transporthttp.StartServer(srv, l.With("pubsub_server", "rest"), errchan)
	l.Infof("running rest server, listening on: %s", srv.Addr)

	select {
	case <-ctx.Done():
		l.Infow("shuting servers down with", "error", ctx.Err().Error())
		transporthttp.ShutdownServer(context.Background(), srv, l.With("server", "rest"))
		postgresEventBus.Close()
	case err := <-errchan:
		l.Info("rest server crashed with ", err.Error())
	}

	//* finaliza execução das subrotinas
	cancel()

	//* espera finalização das subrotinas
	select {
	case <-time.After(20 * time.Second):
		return errors.New("shutdown timedout")
	case <-fanInEvent:
		return nil
	}
}
