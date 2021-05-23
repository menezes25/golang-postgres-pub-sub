package postgres

import (
	"context"
	gopostgrespubsub "postgres_pub_sub"
	"time"
)

func (pc *PostgresClient) StartListeningToNotifications(ctx context.Context, chans []string) error {
	poolConn, err := pc.connPool.Acquire(ctx)
	if err != nil {
		pc.log.Error("acquire connection error: ", err.Error())
		return err
	}

	//* registra a conexão para o channel 'name'
	conn := poolConn.Conn()
	for _, ch := range chans {
		_, err = conn.Exec(ctx, "listen "+ch)
		if err != nil {
			pc.log.Warnw("failed to listen", "error", err.Error(), "channel", ch)
		}
		pc.log.Infow("listening with success", "channel", ch)
		pc.activeChannels = append(pc.activeChannels, ch)
	}

	go func() {
		defer func() {
			pc.log.Debug("stoping event listening")
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			for _, ch := range pc.activeChannels {
				_, err := conn.Exec(ctx, "unlisten "+ch)
				if err != nil {
					pc.log.Error("failed to unlisten event with: ", err.Error())
				}
			}

			poolConn.Release() //! porque se não a conexão com o banco nunca será fechada
			cancel()
		}()

		for {
			//* if ctx is done, err will be non-nil and this func will return
			msg, err := conn.WaitForNotification(ctx)
			if err != nil {
				pc.log.Info("stoped listening to postgres notifications with: ", err.Error())
				return
			}

			pc.log.Infow("new notification from postrges", "channel", msg.Channel, "payload", msg.Payload)
			err = pc.eventBus.Publish(msg.Channel, msg.Payload)
			if err != nil {
				pc.log.Warnw("failed to publish message", "error", err.Error(), "channel", msg.Channel, "message", msg.Payload)
			}
		}
	}()

	return nil
}

func (pc *PostgresClient) ListenToEvents(ctx context.Context, name string) (chan gopostgrespubsub.Notification, error) {
	poolConn, err := pc.connPool.Acquire(ctx)
	if err != nil {
		pc.log.Error("Acquire connection error: ", err.Error())
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
			pc.log.Debug("stoping event listening")
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			_, err := conn.Exec(ctx, "unlisten "+name)
			if err != nil {
				pc.log.Error("failed to unlisten event with: ", err.Error())
			}

			poolConn.Release() //! porque se não a conexão com o banco nunca será fechada
			close(notchan)
			cancel()
		}()

		for {
			//* if ctx is done, err will be non-nil and this func will return
			msg, err := conn.WaitForNotification(ctx)
			if err != nil {
				pc.log.Info("stoped listening to postgres notifications with: ", err.Error())
				return
			}

			pc.log.Infow("new notification from postrges", "channel", msg.Channel, "payload", msg.Payload)

			notchan <- gopostgrespubsub.Notification{Channel: msg.Channel, Payload: msg.Payload}
		}
	}()

	return notchan, nil
}
