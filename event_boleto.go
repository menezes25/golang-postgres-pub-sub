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

type BoletoPayloadRes struct {
	Data      BoletoPayload `json:"data"`
	Success   bool          `json:"success"`
	Operation string        `json:"operation"`
	Message   string        `json:"message"`
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
	fmt.Println("ok")

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

func (e BoletoPayload) handleInsertEvent() (BoletoPayloadRes, error) {
	println("processing insert boleto ", e.Code)
	<-time.After(2000 * time.Millisecond)
	pr := BoletoPayloadRes{
		Data:      e,
		Success:   true,
		Operation: PG_INSERT_OP,
		Message:   fmt.Sprint("boleto ", e.Code, " insert processed"),
	}
	return pr, nil
}

func (e BoletoPayload) handleUpdateEvent() (BoletoPayloadRes, error) {
	println("processing update boleto ", e.Code)
	<-time.After(2000 * time.Millisecond)
	println("boleto", e.Code, "update processed")
	pr := BoletoPayloadRes{
		Data:      e,
		Success:   true,
		Operation: PG_UPDATE_OP,
		Message:   fmt.Sprint("boleto ", e.Code, " update processed"),
	}
	return pr, nil
}

func (e BoletoPayload) handleDeleteEvent() (BoletoPayloadRes, error) {
	println("processing delete boleto ", e.Code)
	<-time.After(2000 * time.Millisecond)
	println("boleto", e.Code, "delete processed")
	pr := BoletoPayloadRes{
		Data:      e,
		Success:   true,
		Operation: PG_DELETE_OP,
		Message:   fmt.Sprint("boleto ", e.Code, " delete processed"),
	}
	return pr, nil
}
