package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"postgres_pub_sub/postgres"
	"strconv"
	"syscall"

	"github.com/jackc/pgx"
	"go.uber.org/zap"
)

func main() {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("failed to get logger with", err.Error())
	}
	logger := zapLogger.Sugar()

	if err := run(logger, &postgres.PostgresClient{}); err != nil {
		logger.Fatal(err.Error())
	}
}

func run(l *zap.SugaredLogger, postgresClient *postgres.PostgresClient) error {
	ctx := context.TODO()

	_, err := startPostgres(ctx, l)
	if err != nil {
		log.Fatalln(err)
	}

	doneChan := make(chan os.Signal, 1)
	signal.Notify(doneChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	oscall := <-doneChan
	l.Infow("shuting servers down with", "sys_signal", oscall.String())

	return err
}

func envPGConfig() pgx.ConnConfig {
	port, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	return pgx.ConnConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     uint16(port),
		Database: os.Getenv("DB_DATABASE"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	}
}

type Payload struct {
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type notification struct {
	Op      string  `json:"op,omitempty"`
	Payload Payload `json:"payload,omitempty"`
}

func startPostgres(ctx context.Context, l *zap.SugaredLogger) (*pgx.ConnPool, error) {
	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: envPGConfig(),
		AfterConnect: func(c *pgx.Conn) error {
			// this subscribes our connection to the 'event' channel, the channel name can be whatever you want.
			err := c.Listen("event")
			if err != nil {
				return err
			}

			go func() {
				for {
					// if ctx is done, err will be non-nil and this func will return
					msg, err := c.WaitForNotification(ctx)
					if err != nil {
						l.Error("WaitForNotification error: ", err.Error())
						return
					}
					meta := &notification{}
					err = json.Unmarshal([]byte(msg.Payload), &meta)
					if err != nil {
						l.Error("json.Unmarshal error: ", err.Error())
						return
					}

					l.Info("Got a message from postgres!!!", "operation", meta.Op, "payload", meta.Payload)
				}
			}()

			return nil
		},
	})
	return pool, err
}
