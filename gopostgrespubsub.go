package gopostgrespubsub

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/menezes25/golang-postgres-pub-sub/eventbus"
	"github.com/menezes25/golang-postgres-pub-sub/postgres"
	transporthttp "github.com/menezes25/golang-postgres-pub-sub/transport/http"
	"go.uber.org/zap"
)

type PostgresPubSub interface {
	StartServer(ctx context.Context) <-chan error
}

type PostgresPubSubConfig struct {
	DbHost        string
	DbName        string
	DbUser        string
	DbPort        string
	DbPass        string
	ServerAddr    string
	EventHandlers []EventHandler
}

// EventHandler é uma interface que deve ser atendida para acessar os triggers
type EventHandler interface {
	Name() string // O nome retornado é o nome da tabela que é também o nome do topico
	// HandleInsert recebe uma string json do tipo chave=coluna_tabela e valor=valor_coluna_tabela
	HandleInsert(string) (interface{}, error)
	// HandleUpdate recebe uma string json do tipo chave=coluna_tabela e valor=valor_coluna_tabela
	HandleUpdate(string) (interface{}, error)
	// HandleDelete recebe uma string json do tipo chave=coluna_tabela e valor=valor_coluna_tabela
	HandleDelete(string) (interface{}, error)
	// HandlerUnknownOperation recebe uma string json do tipo chave=coluna_tabela e valor=valor_coluna_tabela
	HandlerUnknownOperation(string) (interface{}, error)
}

type postgresPubSub struct {
	server *http.Server
	logger *zap.Logger
}

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

	postgresCli, err := postgres.NewProduction(config.DbHost, config.DbName, config.DbUser, config.DbPort, config.DbPass, l)
	if err != nil {
		return nil, err
	}

	pubSubTableNames := make([]string, 0)
	for _, evHandler := range config.EventHandlers {
		pubSubTableNames = append(pubSubTableNames, evHandler.Name())
	}
	postgresEventBus := eventbus.NewEventBus(pubSubTableNames)
	postgresCli.WithEventBus(postgresEventBus)

	err = postgresCli.StartListeningToNotifications(ctx)
	if err != nil {
		return nil, err
	}

	eventChannels := make([]<-chan eventbus.Topic, 0)
	for _, evHandler := range config.EventHandlers {
		topicName := evHandler.Name()
		dataChan := postgresEventBus.Subscribe(topicName)
		eventChan := HandleEvent(ctx, dataChan, l, evHandler)
		eventChannels = append(eventChannels, eventChan)
	}

	fanInEvent := eventbus.Merge(eventChannels...)

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

// DefaultEventHandler implementa gopostgrespubsub.EventHandler
// e por padrão é o handler do topico `event`.
//
// Como usar:
// Use a Composição em Golang e
// sobreescreva o metodo Name() padrão pelo seu metodo Name().
//
// type MyEventHandler struct {
// 	gopostgrespubsub.DefaultEventHandler // composition
// }
//
// func (myEventHandler *MyEventHandler) Name() string {
// 	return "anothertopic"
// }
//
// check if the type satisfies the interface `gopostgrespubsub.EventHandler`
// var _ gopostgrespubsub.EventHandler = (*MyEventHandler)(nil)
//
type DefaultEventHandler struct{}

var _ EventHandler = (*DefaultEventHandler)(nil)

func (b *DefaultEventHandler) Name() string {
	return "event"
}

func (b *DefaultEventHandler) HandleInsert(payload string) (interface{}, error) {
	fmt.Printf("[WARN] DefaultEventHandler INSERT | payload: %v | returning not implemented.\n", payload)
	return nil, errors.New("not implemented")
}

func (b *DefaultEventHandler) HandleUpdate(payload string) (interface{}, error) {
	fmt.Printf("[WARN] DefaultEventHandler UPDATE | payload: %v | returning not implemented.\n", payload)
	return nil, errors.New("not implemented")
}

func (b *DefaultEventHandler) HandleDelete(payload string) (interface{}, error) {
	fmt.Printf("[WARN] DefaultEventHandler DELETE | payload: %v | returning not implemented.\n", payload)
	return nil, errors.New("not implemented")
}

func (b *DefaultEventHandler) HandlerUnknownOperation(payload string) (interface{}, error) {
	fmt.Printf("[WARN] DefaultEventHandler UNKNOWN OPERATION | payload: %v | returning not implemented.\n", payload)
	return nil, errors.New("not implemented")
}
