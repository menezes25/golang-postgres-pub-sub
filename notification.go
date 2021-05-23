package gopostgrespubsub

type EventType string

const (
	EVENT EventType = "event"
)

type Event struct {
	Type    EventType `json:"type,omitempty"`
	Payload string    `json:"payload,omitempty"`
}

type Notification struct {
	Channel string `json:"channel,omitempty"`
	Payload string `json:"payload,omitempty"`
}
