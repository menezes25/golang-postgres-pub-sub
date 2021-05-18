package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

// PostgresClient cliente para manipulações em banco de dados postgres
type PostgresClient struct {
	dns      string
	logger   *zap.SugaredLogger
	connPool *pgxpool.Pool
	psql     squirrel.StatementBuilderType
}

// NewProduction se conecta ao banco de dados tenta criar o banco e suas tabelas e retorna um cliente
func NewProduction(host, dbname, user, port, password string, logger *zap.SugaredLogger) (*PostgresClient, error) {
	//* cliente conectado ao banco de dados padrão para criação do banco de dados
	cli, err := newPostgresClient(host, "postgres", user, port, password, logger)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = cli.connPool.Exec(ctx, createDatabase)
	if err != nil {
		sqlErr := &pgconn.PgError{}
		if errors.As(err, &sqlErr) {
			//* se o banco já existe não queremos retornar erro
			if !isDuplicateDatabaseError(sqlErr) {
				cli.Close()
				return nil, err
			}
		}
	}
	cli.Close()

	cli, err = newPostgresClient(host, dbname, user, port, password, logger)
	if err != nil {
		return nil, err
	}

	err = cli.migrateTables(ctx)
	if err != nil {
		return nil, err
	}

	return cli, nil
}

// newPostgresClient cria um cliente capaz de se comunicar com o banco de dados
func newPostgresClient(host, dbname, user, port, password string, logger *zap.SugaredLogger) (*PostgresClient, error) {
	dns := fmt.Sprintf("host=%s dbname=%s user=%s port=%s password=%s sslmode=disable",
		host,
		dbname,
		user,
		port,
		password,
	)

	logger.Debug("database dns: ", dns)

	config, err := pgxpool.ParseConfig(dns)
	if err != nil {
		return nil, err
	}

	dbpool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		logger.Error("failed to connect with database with: ", err.Error())
		return nil, err
	}

	err = dbpool.Ping(context.Background())
	if err != nil {
		logger.Error("failed to ping database with: ", err.Error())
		return nil, err
	}

	logger.Info("connected to database with success")

	return &PostgresClient{
		dns:      dns,
		logger:   logger,
		connPool: dbpool,
		psql:     squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

func (pc *PostgresClient) Close() {
	pc.connPool.Close()
}

func (pc *PostgresClient) migrateTables(ctx context.Context) error {
	err := pc.createTable(ctx, createEventTable)
	if err != nil {
		return err
	}

	err = pc.createTriggerAndTriggerFuncs(ctx)
	if err != nil {
		return err
	}

	pc.logger.Info("created all tables with success")

	return nil
}

func (pc *PostgresClient) ListenToEvents(ctx context.Context) (chan string, error) {
	notchan := make(chan string, 1)

	poolConn, err := pc.connPool.Acquire(ctx)
	if err != nil {
		pc.logger.Error("Acquire connection error: ", err.Error())
		close(notchan)
		return notchan, err
	}
	
	// this subscribes our connection to the 'event' channel, the channel name can be whatever you want.
	conn := poolConn.Conn()
	_, err = conn.Exec(ctx, "listen event")
	if err != nil {
		close(notchan)
		return notchan, err
	}

	go func() {
		defer func() {
			pc.logger.Debug("stoping event listening")
			_, err := conn.Exec(ctx, "unlisten event")
			if err != nil {
				pc.logger.Error("failed to unlisten event with: ", err.Error())
			}
			poolConn.Release()//!IMPORTANTE
			close(notchan)
		}()
		
		// if ctx is done, err will be non-nil and this func will return
		msg, err := conn.WaitForNotification(ctx)
		if err != nil {
			pc.logger.Error("WaitForNotification error: ", err.Error())
			return
		}

		pc.logger.Infow("Got a message from postgres!!!", "payload", msg.Payload)
		notchan <- msg.Payload
	}()

	return notchan, nil
}

func isDuplicateTableError(sqlErr *pgconn.PgError) bool {
	// erro de tabela já criada no banco de dados
	return sqlErr.Code == "42P07"
}

func isUniqueViolationError(sqlErr *pgconn.PgError) bool {
	// erro de chave única já existente
	return sqlErr.Code == "23505"
}

func isDuplicateDatabaseError(sqlErr *pgconn.PgError) bool {
	// erro de banco de dados já existente no banco de dados
	return sqlErr.Code == "42P04"
}
