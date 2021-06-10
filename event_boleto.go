package gopostgrespubsub

import (
	"fmt"
	"time"
)

//! Tipo de dados vem do usuario da lib
type BoletoPayload struct {
	ID        int       `json:"id,omitempty"`
	Code      string    `json:"code,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

//! Tipo de dados vem do usuario da lib
type BoletoPayloadRes struct {
	Data      BoletoPayload `json:"data"`
	Success   bool          `json:"success"`
	Operation string        `json:"operation"`
	Message   string        `json:"message"`
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
