package gopostgrespubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/menezes25/golang-postgres-pub-sub/internal"
	"go.uber.org/zap"
)

//* Descreve as operações
const (
	PG_INSERT_OP = "INSERT"
	PG_UPDATE_OP = "UPDATE"
	PG_DELETE_OP = "DELETE"
)

type EventHandler interface {
	Name() string
	HandleInsert(interface{}) (interface{}, error)
	HandleUpdate(interface{}) (interface{}, error)
	HandleDelete(interface{}) (interface{}, error)
}

//TODO: trocar dataTopic & handleEvent por EventHandler
//TODO: trocar nome da função para HandleEvent
func HandleData(ctx context.Context, eventDataChan chan internal.DataEvent, l *zap.SugaredLogger, dataTopic string, handleEvent func(eventData internal.DataEvent, responseChan chan internal.Event)) <-chan internal.Event {
	ch := make(chan internal.Event, 1)

	go func() {
		for {
			select {
			case data := <-eventDataChan:
				//TODO: trocar dataTopic por EventHandler.Name()
				if data.Topic == dataTopic {
					l.Infow("new data event", "topic", data.Topic, "payload", data.Data)
					//TODO: trocar handleEvent, por switch & EventHandler.HandleInsert/Update/Delete
					handleEvent(data, ch)
					//* precisamos extrair somente a operação executada
					//* o unmarshal do payload é de responsabilidade do usuário da lib
					type EventData struct {
						Op string `json:"op,omitempty"`
					}

					var event EventData
					payload := data.Data.(string)
					err := json.Unmarshal([]byte(payload), &event)
					if err != nil {
						ch <- internal.Event{Payload: err.Error(), Type: internal.EventType(data.Topic)}
					}
					fmt.Println("ok")

					switch event.Op {
					case PG_INSERT_OP:
						//TODO: chamar EventHandler.HandleInsert
						r, err := event.Payload.handleInsertEvent()
						if err != nil {
							pErr := "error" + err.Error()
							ch <- internal.Event{Payload: pErr, Type: internal.EventType(data.Topic)}
							return
						}
						ch <- internal.Event{Payload: r, Type: internal.EventType(data.Topic)}
					case PG_UPDATE_OP:
						//TODO: chamar EventHandler.HandleUpdate
						r, err := event.Payload.handleUpdateEvent()
						if err != nil {
							pErr := "error" + err.Error()
							ch <- internal.Event{Payload: pErr, Type: internal.EventType(data.Topic)}
							return
						}
						ch <- internal.Event{Payload: r, Type: internal.EventType(data.Topic)}
					case PG_DELETE_OP:
						//TODO: chamar EventHandler.HandleDelete
						r, err := event.Payload.handleDeleteEvent()
						if err != nil {
							pErr := "error" + err.Error()
							ch <- internal.Event{Payload: pErr, Type: internal.EventType(data.Topic)}
							return
						}
						ch <- internal.Event{Payload: r, Type: internal.EventType(data.Topic)}
					default:
						ch <- internal.Event{Payload: "operation unkown"}
					}
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

type DefaultEventHandler struct {
	//* dependencias
}

func (b DefaultEventHandler) Name() string {
	return "boleto"
}

func (b DefaultEventHandler) HandleInsert(payload BoletoPayload) (BoletoPayloadRes, error) {
	return BoletoPayloadRes{}, errors.New("not implemented")
}

func (b DefaultEventHandler) HandleUpdate(payload BoletoPayload) (BoletoPayloadRes, error) {
	return BoletoPayloadRes{}, errors.New("not implemented")
}

func (b DefaultEventHandler) HandleDelete(payload BoletoPayload) (BoletoPayloadRes, error) {
	return BoletoPayloadRes{}, errors.New("not implemented")
}
