package postgres

import (
	"context"
	"errors"

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

	pc.logger.Debug("created table with success")

	return nil
}

func (pc *PostgresClient) createTriggerAndTriggerFuncs(ctx context.Context) error {
	_, err := pc.connPool.Exec(ctx, createTriggerAndTriggerFuncsQry)
	if err != nil {
		return err
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

const createTriggerAndTriggerFuncsQry = `
DO $$
BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'event_notify') THEN

        CREATE OR REPLACE FUNCTION event_notify() RETURNS trigger AS $FN$
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

            PERFORM pg_notify('event', payload::TEXT);
            
            RETURN NEW;

        END;
        $FN$ LANGUAGE plpgsql;

	END IF;

	IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'event_notify_update') THEN
		CREATE TRIGGER event_notify_update AFTER UPDATE ON event FOR EACH ROW EXECUTE PROCEDURE event_notify();
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'event_notify_insert') THEN
		CREATE TRIGGER event_notify_insert AFTER INSERT ON event FOR EACH ROW EXECUTE PROCEDURE event_notify();
	END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'event_notify_delete') THEN
		CREATE TRIGGER event_notify_delete AFTER DELETE ON event FOR EACH ROW EXECUTE PROCEDURE event_notify();
	END IF;
END $$;
`
