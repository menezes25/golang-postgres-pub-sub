package postgres

import (
	"context"
	"time"
)

const (
	eventTable = "event"
)

func (pc *PostgresClient) AddEvent(ctx context.Context, name string) error {
	qry, _, err := pc.psql.Insert(eventTable).Columns("name", "created_at").Values(1, 2).ToSql()
	if err != nil {
		return err
	}

	_, err = pc.connPool.Exec(ctx, qry, name, time.Now())
	if err != nil {
		return err
	}

	pc.log.Debugw("added event with success", "name", name)

	return nil
}

func (pc *PostgresClient) UpdateEventByID(ctx context.Context, name string, id int) error {
	qry, _, err := pc.psql.Update(eventTable).Set("name", 1).Set("updated_at", 2).Where("id=?").ToSql()
	if err != nil {
		return err
	}

	_, err = pc.connPool.Exec(ctx, qry, name, time.Now(), id)
	if err != nil {
		return err
	}

	pc.log.Debugw("updated event with success")

	return nil
}

func (pc *PostgresClient) DeleteEventByID(ctx context.Context, id int) error {
	qry, _, err := pc.psql.Delete(eventTable).Where("id=?").ToSql()
	if err != nil {
		return err
	}

	_, err = pc.connPool.Exec(ctx, qry, id)
	if err != nil {
		return err
	}

	pc.log.Debugw("updated event with success")

	return nil
}
