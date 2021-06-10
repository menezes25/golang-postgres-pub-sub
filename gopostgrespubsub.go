package gopostgrespubsub

import (
	"context"
	"fmt"
	"net/http"

	"github.com/menezes25/golang-postgres-pub-sub/internal"
	"github.com/menezes25/golang-postgres-pub-sub/postgres"
	transporthttp "github.com/menezes25/golang-postgres-pub-sub/transport/http"
	"go.uber.org/zap"
)

type PostgresPubSub interface {
	StartServer(ctx context.Context) <-chan error
}

type PostgresPubSubConfig struct {
	DbHost           string
	DbName           string
	DbUser           string
	DbPort           string
	DbPass           string
	ServerAddr       string
	PubSubTableNames []string // Os nomes das tableas são também os nomes dos topicos
}

type postgresPubSub struct {
	server *http.Server
	logger *zap.Logger
}

// type satisfies an interface
var _ PostgresPubSub = (*postgresPubSub)(nil)

func NewPostgresPubSub(ctx context.Context, config PostgresPubSubConfig) (PostgresPubSub, error) {
	// FIXME: Verificar essa dependencia
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.OutputPaths = []string{"stdout"}
	zapLogger, err := loggerConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to get logger with %w", err)
	}
	l := zapLogger.Sugar()

	// ctx, cancel := context.WithCancel(context.Background())
	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	// defer cancel()

	postgresCli, err := postgres.NewProduction(config.DbHost, config.DbName, config.DbUser, config.DbPort, config.DbPass, l)
	if err != nil {
		return nil, err
	}

	// const topicEvent = "event"
	// const topicBoleto = "boleto"
	// postgresEventBus := gopostgrespubsub.NewEventBus([]string{topicEvent, topicBoleto})
	postgresEventBus := internal.NewEventBus(config.PubSubTableNames)
	postgresCli.WithEventBus(postgresEventBus)

	err = postgresCli.StartListeningToNotifications(ctx)
	if err != nil {
		return nil, err
	}

	eventChannels := make([]<-chan internal.Event, 0)
	for _, tableName := range config.PubSubTableNames {
		topicName := tableName
		dataChan := postgresEventBus.Subscribe(topicName)
		// FIXME: O Handle deve ser injetado
		eventChan := internal.HandleData(ctx, dataChan, l, topicName, HandlePostgresDataBoleto)
		eventChannels = append(eventChannels, eventChan)
	}

	fanInEvent := internal.Merge(eventChannels...)

	transHttpRouter, err := transporthttp.NewTransportHttp(ctx, postgresCli, fanInEvent)
	if err != nil {
		return nil, err
	}

	srv := transporthttp.NewServer(config.ServerAddr, transHttpRouter)
	postgresPS := postgresPubSub{
		server: srv,
		logger: zapLogger,
	}

	// Quando o context finalizar fecha o postgresEventBus
	go func() {
		select {
		case <-ctx.Done():
			postgresEventBus.Close()
		}
	}()

	return &postgresPS, nil
}

func (ps *postgresPubSub) StartServer(ctx context.Context) <-chan error {
	l := ps.logger.Sugar()
	errChanServer := make(chan error)
	go func() {
		errchan := make(chan error, 1)
		go transporthttp.StartServer(ps.server, l.With("pubsub_server", "rest"), errchan)
		l.Infof("running rest server, listening on: %s", ps.server.Addr)

		select {
		case <-ctx.Done():
			l.Infow("shuting servers down with", "error", ctx.Err().Error())
			transporthttp.ShutdownServer(context.Background(), ps.server, l.With("server", "rest"))

		case err := <-errchan:
			l.Info("rest server crashed with ", err.Error())
			errChanServer <- err
		}
	}()
	return errChanServer
}
