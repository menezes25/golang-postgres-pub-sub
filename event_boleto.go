package gopostgrespubsub

import (
	"encoding/json"
	"fmt"
	"time"
)

type BoletoPayload struct {
	ID        int       `json:"id,omitempty"`
	Code      string    `json:"code,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func HandlePostgresDataBoleto(eventData DataEvent, responseChan chan Event) {
	type EventData struct {
		Op      string        `json:"op,omitempty"`
		Payload BoletoPayload `json:"payload,omitempty"`
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

func (e BoletoPayload) handleInsertEvent() (string, error) {
	println("processing insert boleto ", e.Code)
	<-time.After(2000 * time.Millisecond)
	return fmt.Sprint("boleto ", e.Code, " insert processed"), nil
}

func (e BoletoPayload) handleUpdateEvent() (string, error) {
	println("processing update boleto ", e.Code)
	<-time.After(2000 * time.Millisecond)
	println("boleto", e.Code, "update processed")
	return "", nil
}

func (e BoletoPayload) handleDeleteEvent() (string, error) {
	println("processing delete boleto ", e.Code)
	<-time.After(2000 * time.Millisecond)
	println("boleto", e.Code, "delete processed")
	return "", nil
}
