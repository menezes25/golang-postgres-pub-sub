package postgres

import (
	"context"
	gopostgrespubsub "postgres_pub_sub"
	"time"
)

func (pc *PostgresClient) ListenToEvents(ctx context.Context, name string) (chan gopostgrespubsub.Notification, error) {
	poolConn, err := pc.connPool.Acquire(ctx)
	if err != nil {
		pc.logger.Error("Acquire connection error: ", err.Error())
		return nil, err
	}

	//* registra a conexão para o channel 'name'
	conn := poolConn.Conn()
	_, err = conn.Exec(ctx, "listen "+name)
	if err != nil {
		return nil, err
	}

	notchan := make(chan gopostgrespubsub.Notification, 1)
	go func() {
		defer func() {
			pc.logger.Debug("stoping event listening")
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			_, err := conn.Exec(ctx, "unlisten "+name)
			if err != nil {
				pc.logger.Error("failed to unlisten event with: ", err.Error())
			}

			poolConn.Release() //! porque se não a conexão com o banco nunca será fechada
			close(notchan)
			cancel()
		}()

		for {
			//* if ctx is done, err will be non-nil and this func will return
			msg, err := conn.WaitForNotification(ctx)
			if err != nil {
				pc.logger.Error("WaitForNotification error: ", err.Error())
				return
			}

			pc.logger.Infow("new notification from postrges", "channel", msg.Channel, "payload", msg.Payload)

			notchan <- gopostgrespubsub.Notification{Channel: msg.Channel, Payload: msg.Payload}
		}
	}()

	return notchan, nil
}
