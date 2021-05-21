package gopostgrespubsub

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

const (
	EVENT_INSERT_OP = "INSERT"
	EVENT_UPDATE_OP = "UPDATE"
	EVENT_DELETE_OP = "DELETE"
)

type notification struct {
	Op    string `json:"op,omitempty"`
	Event Event  `json:"payload,omitempty"`
}

type Event struct {
	ID        int       `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func HandleEventEvents(notchan chan string, l *zap.SugaredLogger) chan string {
	ch := make(chan string, 1)

	go func() {
		for newNotification := range notchan {
			l.Infow("new notification", "payload", newNotification)
			var not notification
			err := json.Unmarshal([]byte(newNotification), &not)
			if err != nil {
				l.Errorw("failed to unmarshal notification", "error", err.Error())
			}
			switch not.Op {
			case EVENT_INSERT_OP:
				r, err := not.Event.handleInsertEvent()
				if err != nil {
					r = "error" + err.Error()
				}
				ch <- r
			case EVENT_UPDATE_OP:
				r, err := not.Event.handleUpdateEvent()
				if err != nil {
					r = "error" + err.Error()
				}
				ch <- r
			case EVENT_DELETE_OP:
				r, err := not.Event.handleDeleteEvent()
				if err != nil {
					r = "error" + err.Error()
				}
				ch <- r
			}
		}
		l.Info("stoped listening to notification channel")
		close(ch)
	}()

	return ch
}

func (e Event) handleInsertEvent() (string, error) {
	println("processing insert event ", e.Name)
	<-time.After(500 * time.Millisecond)
	return fmt.Sprint("event ", e.Name, " insert processed"), nil
}

func (e Event) handleUpdateEvent() (string, error) {
	println("processing update event ", e.Name)
	<-time.After(500 * time.Millisecond)
	println("event", e.Name, "update processed")
	return "", nil
}

func (e Event) handleDeleteEvent() (string, error) {
	println("processing delete event ", e.Name)
	<-time.After(500 * time.Millisecond)
	println("event", e.Name, "delete processed")
	return "", nil
}
