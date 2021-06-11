package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/menezes25/golang-postgres-pub-sub/eventbus"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

// PostgresClient cliente para manipulações em banco de dados postgres
type PostgresClient struct {
	log *zap.SugaredLogger

	dns            string
	connPool       *pgxpool.Pool
	activeChannels []string // armazena canais de listen/notify do postgres

	psql squirrel.StatementBuilderType

	eventBus *eventbus.EventBus
}

// NewProduction se conecta ao banco de dados tenta criar o banco e suas tabelas e retorna um cliente
func NewProduction(host, dbname, user, port, password string, logger *zap.SugaredLogger) (*PostgresClient, error) {
	//* cliente conectado ao banco de dados padrão para criação do banco de dados
	cli, err := newPostgresClient(host, "postgres", user, port, password, logger)
	if err != nil {
		return nil, err
	}

	cli.activeChannels = make([]string, 0)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// TODO: Como LIB não se cria database
	// _, err = cli.connPool.Exec(ctx, createDatabase)
	// if err != nil {
	// 	sqlErr := &pgconn.PgError{}
	// 	errors.As(err, &sqlErr)
	// 	//* se o banco já existe não queremos retornar erro
	// 	if !isDuplicateDatabaseError(sqlErr) {
	// 		cli.Close()
	// 		return nil, err
	// 	}

	// 	// cli.Close()
	// 	// return nil, err
	// }
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
	dns := fmt.Sprintf("host=%s dbname=%s user=%s port=%s password=%s sslmode=disable application_name=pub-sub-go",
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
		log:      logger,
		connPool: dbpool,
		psql:     squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

func (pc *PostgresClient) Close() {
	pc.connPool.Close()
}

func (pc *PostgresClient) WithEventBus(eventBus *eventbus.EventBus) {
	pc.eventBus = eventBus
}

func (pc *PostgresClient) migrateTables(ctx context.Context) error {
	// err := pc.createTable(ctx, createEventTable)
	// if err != nil {
	// 	return err
	// }

	// err = pc.createTable(ctx, createBoletoTable)
	// if err != nil {
	// 	return err
	// }

	// err := pc.createTriggerAndTriggerFuncs(ctx, []string{"event", "boleto"})
	err := pc.createTriggerAndTriggerFuncs(ctx, []string{"boleto"})
	if err != nil {
		return err
	}

	pc.log.Info("created all tables with success")

	return nil
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
