package gopostgrespubsub

import (
	"context"
	"encoding/json"

	"github.com/menezes25/golang-postgres-pub-sub/eventbus"
	"go.uber.org/zap"
)

//* Descreve as operações
const (
	PG_INSERT_OP = "INSERT"
	PG_UPDATE_OP = "UPDATE"
	PG_DELETE_OP = "DELETE"
)

func HandleEvent(ctx context.Context, eventDataChan chan eventbus.DataEvent, l *zap.SugaredLogger, evHandler EventHandler) <-chan eventbus.Topic {
	ch := make(chan eventbus.Topic, 1)

	go func() {
		dataTopic := evHandler.Name()
		for {
			select {
			case data := <-eventDataChan:

				if data.Topic != dataTopic {
					l.Warnw("obtained postgres event with different topic than registered.", "topic", dataTopic, "event", data.Data)
					continue
				}

				l.Infow("new data event", "topic", data.Topic, "payload", data.Data)
				type EventData struct {
					Op      string                 `json:"op,omitempty"`
					Payload map[string]interface{} `json:"payload,omitempty"`
				}

				var event EventData
				payload := data.Data.(string)
				err := json.Unmarshal([]byte(payload), &event)
				if err != nil {
					l.Errorf("EventData Unmarshal %s", err)
					ch <- eventbus.Topic{Payload: err.Error(), Type: data.Topic}
					continue
				}

				payloadByteArr, err := json.Marshal(event.Payload)
				if err != nil {
					l.Errorf("Payload Marshal %s", err)
					ch <- eventbus.Topic{Payload: err.Error(), Type: data.Topic}
					continue
				}

				var handleRes interface{}
				var handleErr error
				switch event.Op {
				case PG_INSERT_OP:
					handleRes, handleErr = evHandler.HandleInsert(string(payloadByteArr))
				case PG_UPDATE_OP:
					handleRes, handleErr = evHandler.HandleUpdate(string(payloadByteArr))
				case PG_DELETE_OP:
					handleRes, handleErr = evHandler.HandleDelete(string(payloadByteArr))
				default:
					handleRes, handleErr = evHandler.HandlerUnknownOperation(string(payloadByteArr))
				}

				// Caso handleErr && handleRes forem nulos, não escreve nada no canal

				if handleErr != nil {
					ch <- eventbus.Topic{Payload: handleErr.Error(), Type: data.Topic}
					continue
				}

				if handleRes != nil {
					ch <- eventbus.Topic{Payload: handleRes, Type: data.Topic}
					continue
				}

			case <-ctx.Done():
				l.Infof("stop handling eventDataChan. topic: %s", dataTopic)
				close(ch)
				return
			}
		}
	}()

	return ch
}
