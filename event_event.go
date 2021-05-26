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

type EventPayloadRes struct {
	Data      EventPayload `json:"data"`
	Success   bool         `json:"success"`
	Operation string       `json:"operation"`
	Message   string       `json:"message"`
}

func HandlePostgresDataEvent(eventData DataEvent, responseChan chan Event) {
	type EventData struct {
		Op      string       `json:"op,omitempty"`
		Payload EventPayload `json:"payload,omitempty"`
	}

	var event EventData
	payload := eventData.Data.(string)
	err := json.Unmarshal([]byte(payload), &event)
	if err != nil {
		responseChan <- Event{Payload: err.Error(), Type: EventType(eventData.Topic)}
	}

	switch event.Op {
	case PG_INSERT_OP:
		r, err := event.Payload.handleInsertEvent()
		if err != nil {
			pErr := "error" + err.Error()
			responseChan <- Event{Payload: pErr, Type: EventType(eventData.Topic)}
			return
		}
		responseChan <- Event{Payload: r, Type: EventType(eventData.Topic)}
	case PG_UPDATE_OP:
		r, err := event.Payload.handleUpdateEvent()
		if err != nil {
			pErr := "error" + err.Error()
			responseChan <- Event{Payload: pErr, Type: EventType(eventData.Topic)}
			return
		}
		responseChan <- Event{Payload: r, Type: EventType(eventData.Topic)}
	case PG_DELETE_OP:
		r, err := event.Payload.handleDeleteEvent()
		if err != nil {
			pErr := "error" + err.Error()
			responseChan <- Event{Payload: pErr, Type: EventType(eventData.Topic)}
			return
		}
		responseChan <- Event{Payload: r, Type: EventType(eventData.Topic)}
	default:
		responseChan <- Event{Payload: "operation unkown"}
	}
}

func (e EventPayload) handleInsertEvent() (EventPayloadRes, error) {
	println("processing insert event ", e.Name)
	<-time.After(500 * time.Millisecond)
	bpr := EventPayloadRes{
		Data:      e,
		Success:   true,
		Operation: PG_INSERT_OP,
		Message:   fmt.Sprint("event ", e.Name, " insert processed"),
	}
	return bpr, nil
}

func (e EventPayload) handleUpdateEvent() (EventPayloadRes, error) {
	println("processing update event ", e.Name)
	<-time.After(500 * time.Millisecond)
	println("event", e.Name, "update processed")
	pr := EventPayloadRes{
		Data:      e,
		Success:   true,
		Operation: PG_UPDATE_OP,
		Message:   fmt.Sprint("event ", e.Name, " update processed"),
	}
	return pr, nil
}

func (e EventPayload) handleDeleteEvent() (EventPayloadRes, error) {
	println("processing delete event ", e.Name)
	<-time.After(500 * time.Millisecond)
	println("event", e.Name, "delete processed")
	pr := EventPayloadRes{
		Data:      e,
		Success:   true,
		Operation: PG_DELETE_OP,
		Message:   fmt.Sprint("event ", e.Name, " delete processed"),
	}
	return pr, nil
}
