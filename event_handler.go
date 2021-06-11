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

func HandleEvent(ctx context.Context, eventDataChan chan eventbus.DataEvent, l *zap.SugaredLogger, evHandler EventHandler) <-chan eventbus.Event {
	ch := make(chan eventbus.Event, 1)

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
					Op string `json:"op,omitempty"`
					// TODO: Verificar se essa é a melhor estrutura para payload, pois o unmarshal do payload é de responsabilidade do usuário da lib
					Payload interface{} `json:"payload,omitempty"`
				}

				var event EventData
				payload := data.Data.(string)
				err := json.Unmarshal([]byte(payload), &event)
				if err != nil {
					ch <- eventbus.Event{Payload: err.Error(), Type: eventbus.EventType(data.Topic)}
				}

				var handleRes interface{}
				var handleErr error
				switch event.Op {
				case PG_INSERT_OP:
					handleRes, handleErr = evHandler.HandleInsert(event.Payload)
				case PG_UPDATE_OP:
					handleRes, handleErr = evHandler.HandleUpdate(event.Payload)
				case PG_DELETE_OP:
					handleRes, handleErr = evHandler.HandleDelete(event.Payload)
				default:
					handleRes, handleErr = evHandler.HandlerUnknownOperation(event.Payload)
				}

				// Caso handleErr && handleRes forem nulos, não escreve nada no canal

				if handleErr != nil {
					ch <- eventbus.Event{Payload: handleErr.Error(), Type: eventbus.EventType(data.Topic)}
					continue
				}

				if handleRes != nil {
					ch <- eventbus.Event{Payload: handleRes, Type: eventbus.EventType(data.Topic)}
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
