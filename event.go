package gopostgrespubsub

import (
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

// type Eventer interface {
// 	Listen(EventType)
// 	Unlisten(EventType)
// }

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

func HandleEventEvents(notchan chan Notification, l *zap.SugaredLogger) chan Event {
	ch := make(chan Event, 1)

	go func() {
		for notification := range notchan {
			l.Infow("new notification", "payload", notification)
			switch notification.Channel {
			//TODO: parametrizar canais
			case "event":
				handleEventNotification(notification, ch)
			default:
				l.Debugw("got notification with unregistered channel", "notification", notification)
			}
		}
		l.Info("stoped listening to Notification channel")
		close(ch)
	}()

	return ch
}

func handleEventNotification(notification Notification, responseChan chan Event) {
	var event PgEvent
	err := json.Unmarshal([]byte(notification.Payload), &event)
	if err != nil {
		responseChan <- Event{Payload: err.Error()}
	}
	switch event.Op {
	case PG_INSERT_OP:
		r, err := event.Payload.handleInsertEvent()
		if err != nil {
			r = "error" + err.Error()
		}
		responseChan <- Event{Payload: r}
	case PG_UPDATE_OP:
		r, err := event.Payload.handleUpdateEvent()
		if err != nil {
			r = "error" + err.Error()
		}
		responseChan <- Event{Payload: r}
	case PG_DELETE_OP:
		r, err := event.Payload.handleDeleteEvent()
		if err != nil {
			r = "error" + err.Error()
		}
		responseChan <- Event{Payload: r}
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
