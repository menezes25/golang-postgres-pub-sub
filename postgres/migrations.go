package postgres

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"text/template"

	"github.com/jackc/pgconn"
)

func (pc *PostgresClient) createTable(ctx context.Context, qry string) error {
	_, err := pc.connPool.Exec(ctx, qry)
	if err != nil {
		sqlErr := &pgconn.PgError{}
		ok := errors.As(err, &sqlErr)
		if !ok {
			return err
		}

		if isDuplicateTableError(sqlErr) {
			return nil
		}

		return err
	}

	pc.log.Debug("created table with success")

	return nil
}

func (pc *PostgresClient) createTriggerAndTriggerFuncs(ctx context.Context, tables []string) error {
	//TODO: parametriar nome dos canais

	for _, table := range tables {
		tableName := strings.TrimSpace(strings.ToLower(table))
		vars := map[string]string{
			"TableName":         tableName,
			"TriggerFuncName":   tableName + "_notify",
			"TriggerInsertName": tableName + "_notify_insert",
			"TriggerUpdateName": tableName + "_notify_update",
			"TriggerDeleteName": tableName + "_notify_delete",
		}

		tmpl, err := template.New("tmpl").Parse(tmplCreateTriggerAndTriggerFuncsQry)
		if err != nil {
			return err
		}

		var tmplBytes bytes.Buffer
		tmpl.Execute(&tmplBytes, vars)
		if err != nil {
			return err
		}

		createTriggerAndTriggerFuncsQry := tmplBytes.String()

		_, err = pc.connPool.Exec(ctx, createTriggerAndTriggerFuncsQry)
		if err != nil {
			return err
		}
	}

	return nil
}

const deleteTestDatabase = `
DROP DATABASE event_test
`

const createTestDatabase = `
CREATE DATABASE event_test
    WITH 
    OWNER = tester
    ENCODING = 'UTF8'
    CONNECTION LIMIT = -1;
`

const createDatabase = `
CREATE DATABASE event
    WITH 
    OWNER = tester
    ENCODING = 'UTF8'
    CONNECTION LIMIT = -1;
`

const createEventTable = `
CREATE TABLE "event" (
	"id" SERIAL PRIMARY KEY ,
	"name" varchar,
	"created_at" timestamp,
	"updated_at" timestamp
  );
`

const createBoletoTable = `
CREATE TABLE "boleto" (
	"id" SERIAL PRIMARY KEY,
	"code" varchar(55) NOT NULL,
	"created_at" timestamp NOT NULL DEFAULT NOW(),
	"updated_at" timestamp NOT NULL DEFAULT NOW()
  );
`

const tmplCreateTriggerAndTriggerFuncsQry = `
DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname = '{{.TriggerFuncName}}') THEN

        CREATE OR REPLACE FUNCTION {{.TriggerFuncName}}() RETURNS trigger AS $FN$
        DECLARE
            payload jsonb;
        BEGIN
			IF TG_OP = 'DELETE' THEN
				payload = jsonb_build_object(
					'op', to_jsonb(TG_OP),
					'payload', to_jsonb(OLD)
				);
			ELSE
				payload = jsonb_build_object(
					'op', to_jsonb(TG_OP),
					'payload', to_jsonb(NEW)
				);
			END IF;

            PERFORM pg_notify('{{.TableName}}', payload::TEXT);
            
            RETURN NEW;

        END;
        $FN$ LANGUAGE plpgsql;

	END IF;

	IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = '{{.TriggerUpdateName}}') THEN
		CREATE TRIGGER {{.TriggerUpdateName}} AFTER UPDATE ON {{.TableName}} FOR EACH ROW EXECUTE PROCEDURE {{.TriggerFuncName}}();
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = '{{.TriggerInsertName}}') THEN
		CREATE TRIGGER {{.TriggerInsertName}} AFTER INSERT ON {{.TableName}} FOR EACH ROW EXECUTE PROCEDURE {{.TriggerFuncName}}();
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = '{{.TriggerDeleteName}}') THEN
		CREATE TRIGGER {{.TriggerDeleteName}} AFTER DELETE ON {{.TableName}} FOR EACH ROW EXECUTE PROCEDURE {{.TriggerFuncName}}();
	END IF;
END $$;
`
