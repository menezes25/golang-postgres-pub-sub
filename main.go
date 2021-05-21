package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"postgres_pub_sub/postgres"
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

type Payload struct {
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type notification struct {
	Op      string  `json:"op,omitempty"`
	Payload Payload `json:"payload,omitempty"`
}

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
}

func run(l *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	postgresCli, err := postgres.NewProduction(db_host, db_name, db_user, db_port, db_pass, l)
	if err != nil {
		return err
	}

	notificationChan, err := postgresCli.ListenToEvents(ctx)
	if err != nil {
		return err
	}

	go func() {
		for newNotification := range notificationChan {
			l.Infow("new notification", "payload", newNotification)
		}
		l.Info("stoped listening to notification channel")
	}()

	doneChan := make(chan os.Signal, 1)
	signal.Notify(doneChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	oscall := <-doneChan
	l.Infow("shuting servers down with", "sys_signal", oscall.String())

	//* finaliza execução das subrotinas
	cancel()

	//* espera finalização das subrotinas
	select {
	case <-time.After(20*time.Second):
		return errors.New("shutdown timedout")
	case <-notificationChan:
		return nil
	}
}
