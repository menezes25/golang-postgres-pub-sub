package gopostgrespubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

//* Descreve as operações
const (
	PG_INSERT_OP = "INSERT"
	PG_UPDATE_OP = "UPDATE"
	PG_DELETE_OP = "DELETE"
)

type PgEvent struct {
	Op      string       `json:"op,omitempty"`
	Payload EventPayload `json:"payload,omitempty"`
}

type EventPayload struct {
	ID        int       `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func HandleEventData(ctx context.Context, eventDataChan chan DataEvent, l *zap.SugaredLogger) <-chan Event {
	ch := make(chan Event, 1)

	go func() {
		for {
			select {
			case data := <- eventDataChan:
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

func handlePostgresDataEvent(eventData DataEvent, responseChan chan Event) {
	var event PgEvent
	payload := eventData.Data.(string)
	err := json.Unmarshal([]byte(payload), &event)
	if err != nil {
		responseChan <- Event{Payload: err.Error(), Type: EventType(eventData.Topic)}
	}
	switch event.Op {
	case PG_INSERT_OP:
		r, err := event.Payload.handleInsertEvent()
		if err != nil {
			r = "error" + err.Error()
		}
		responseChan <- Event{Payload: r, Type: EventType(eventData.Topic)}
	case PG_UPDATE_OP:
		r, err := event.Payload.handleUpdateEvent()
		if err != nil {
			r = "error" + err.Error()
		}
		responseChan <- Event{Payload: r, Type: EventType(eventData.Topic)}
	case PG_DELETE_OP:
		r, err := event.Payload.handleDeleteEvent()
		if err != nil {
			r = "error" + err.Error()
		}
		responseChan <- Event{Payload: r, Type: EventType(eventData.Topic)}
	default:
		responseChan <- Event{Payload: "operation unkown"}
	}
}

func (e EventPayload) handleInsertEvent() (string, error) {
	println("processing insert event ", e.Name)
	<-time.After(500 * time.Millisecond)
	return fmt.Sprint("event ", e.Name, " insert processed"), nil
}

func (e EventPayload) handleUpdateEvent() (string, error) {
	println("processing update event ", e.Name)
	<-time.After(500 * time.Millisecond)
	println("event", e.Name, "update processed")
	return "", nil
}

func (e EventPayload) handleDeleteEvent() (string, error) {
	println("processing delete event ", e.Name)
	<-time.After(500 * time.Millisecond)
	println("event", e.Name, "delete processed")
	return "", nil
}
