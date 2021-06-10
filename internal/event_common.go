package internal

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

func HandleData(ctx context.Context, eventDataChan chan DataEvent, l *zap.SugaredLogger, dataTopic string, handleEvent func(eventData DataEvent, responseChan chan Event)) <-chan Event {
	ch := make(chan Event, 1)

	go func() {
		for {
			select {
			case data := <-eventDataChan:
				if data.Topic == dataTopic {
					l.Infow("new data event", "topic", data.Topic, "payload", data.Data)
					handleEvent(data, ch)
					continue
				}
				l.Warnw("obtained postgres event with different topic than registered.", "topic", dataTopic, "event", data.Data)

			case <-ctx.Done():
				l.Infof("stop handling eventDataChan. topic: %s", dataTopic)
				close(ch)
				return
			}
		}
	}()

	return ch
}
