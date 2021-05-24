package gopostgrespubsub

import (
	"encoding/json"
	"fmt"
	"time"
)

type EventPayload struct {
	ID        int       `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
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
		r, err := event.Payload.(EventPayload).handleInsertEvent()
		if err != nil {
			r = "error" + err.Error()
		}
		responseChan <- Event{Payload: r, Type: EventType(eventData.Topic)}
	case PG_UPDATE_OP:
		r, err := event.Payload.(EventPayload).handleUpdateEvent()
		if err != nil {
			r = "error" + err.Error()
		}
		responseChan <- Event{Payload: r, Type: EventType(eventData.Topic)}
	case PG_DELETE_OP:
		r, err := event.Payload.(EventPayload).handleDeleteEvent()
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
