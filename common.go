package gopostgrespubsub

import (
	"context"

	"go.uber.org/zap"
)

//* Descreve as operações
const (
	PG_INSERT_OP = "INSERT"
	PG_UPDATE_OP = "UPDATE"
	PG_DELETE_OP = "DELETE"
)

func HandleEventData(ctx context.Context, eventDataChan chan DataEvent, l *zap.SugaredLogger) <-chan Event {
	ch := make(chan Event, 1)

	go func() {
		for {
			select {
			case data := <-eventDataChan:
				l.Infow("new data event", "topic", data.Topic, "payload", data.Data)
				switch data.Topic {
				//TODO: parametrizar canais acho que esses canais vão ficar dentro do EventBus
				case "event":
					handlePostgresDataEvent(data, ch)
				default:
					l.Warnw("got postgres event with unregistered channel", "event", data.Data)
				}
			case <-ctx.Done():
				l.Info("stop handling eventDataChan")
				close(ch)
				return
			}
		}
	}()

	return ch
}
