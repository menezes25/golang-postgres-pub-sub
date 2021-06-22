package postgres

// import (
// 	"context"
// 	"errors"
// 	"testing"
// 	"time"

// 	"github.com/jackc/pgconn"
// 	"go.uber.org/zap"
// )

// var (
// 	host     = "localhost"
// 	name     = "event_test"
// 	user     = "tester"
// 	port     = "5432"
// 	password = "tester"
// )

// func NewTestClient(t *testing.T) *PostgresClient {
// 	zap, err := zap.NewDevelopment()
// 	if err != nil {
// 		t.Fatal("falied to create logger: ", err.Error())
// 	}
// 	logger := zap.Sugar()

// 	cli, err := newPostgresClient(host, "postgres", user, port, password, logger)
// 	if err != nil {
// 		t.Fatal("failed to connect to default database with: ", err.Error())
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	_, err = cli.connPool.Exec(ctx, createTestDatabase)
// 	if err != nil {
// 		sqlErr := &pgconn.PgError{}
// 		if errors.As(err, &sqlErr) {
// 			if isDuplicateDatabaseError(sqlErr) {
// 				return nil
// 			}
// 		}
// 		t.Fatal("failed to create test database with: ", err.Error())
// 	}

// 	cli.connPool.Close()

// 	cli, err = newPostgresClient(host, name, user, port, password, logger)
// 	if err != nil {
// 		t.Fatal("failed to create postgres client with: ", err.Error())
// 	}

// 	err = cli.migrateTables(ctx)
// 	if err != nil {
// 		t.Fatal("failed to migrate tables with: ", err.Error())
// 	}

// 	return cli
// }

// func (pc *PostgresClient) DeleteTestDatabase(t *testing.T) {
// 	zap, err := zap.NewDevelopment()
// 	if err != nil {
// 		t.Fatal("falied to create logger: ", err.Error())
// 	}
// 	logger := zap.Sugar()

// 	cli, err := newPostgresClient(host, "postgres", user, port, password, logger)
// 	if err != nil {
// 		t.Fatal("failed to connect to default database with: ", err.Error())
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	_, err = cli.connPool.Exec(ctx, deleteTestDatabase)
// 	if err != nil {
// 		t.Fatal("failed to delete test database with: ", err.Error())
// 	}

// 	cli.connPool.Close()
// }
