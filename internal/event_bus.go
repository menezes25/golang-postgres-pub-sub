package internal

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrEventBusClosed = errors.New("event bus closed")
	ErrTopicNotFound  = errors.New("topic not found")
)

type DataEvent struct {
	Data  interface{}
	Topic string
}

type DataChannel chan DataEvent

type DataChannelSlice []DataChannel

func NewEventBus(topics []string) *EventBus {
	return &EventBus{
		subscribers: make(map[string]DataChannelSlice),
		closed:      false,
		Topics:      topics,
	}
}

type EventBus struct {
	subscribers map[string]DataChannelSlice
	rm          sync.RWMutex
	closed      bool
	Topics      []string
}

// Subscribe registra o eventbus em um tópico e retorna o canal onde as informações serão enviadas
func (eb *EventBus) Subscribe(topic string) DataChannel {
	eb.rm.Lock()
	defer eb.rm.Unlock()

	ch := make(DataChannel, 1)
	if chans, found := eb.subscribers[topic]; found {
		eb.subscribers[topic] = append(chans, ch)
		return ch
	}

	eb.subscribers[topic] = []DataChannel{ch}
	eb.Topics = append(eb.Topics, topic)
	return ch
}

// Unsubscribe remove o tópico da lista de registrados, fecha todos os canais relacionados a esse tópico
func (eb *EventBus) Unsubscribe(topic string) error {
	eb.rm.Lock()
	defer eb.rm.Unlock()

	var topics DataChannelSlice

	chans, found := eb.subscribers[topic]
	if !found {
		return fmt.Errorf("%w %s", ErrTopicNotFound, topic)
	}

	copy(topics, chans)
	delete(eb.subscribers, topic)
	for i, tpc := range eb.Topics {
		if tpc == topic {
			copy(eb.Topics[i:], eb.Topics[i+1:])
			eb.Topics[len(eb.Topics)-1] = ""
			eb.Topics = eb.Topics[:len(eb.Topics)-1]
		}
	}
	go func() {
		for _, ch := range topics {
			close(ch)
		}
	}()

	return nil
}

func (eb *EventBus) Publish(topic string, data interface{}) error {
	eb.rm.RLock()
	defer eb.rm.RUnlock()

	if eb.closed {
		return ErrEventBusClosed
	}

	chans, found := eb.subscribers[topic]
	if !found {
		return ErrTopicNotFound
	}

	topics := make(DataChannelSlice, len(chans))
	copy(topics, chans)
	go func(data DataEvent, dataChans DataChannelSlice) {
		for _, ch := range dataChans {
			ch <- data
		}
	}(DataEvent{Data: data, Topic: topic}, topics)

	return nil
}

func (eb *EventBus) Close() {
	eb.rm.Lock()
	defer eb.rm.Unlock()

	if !eb.closed {
		for _, subs := range eb.subscribers {
			for _, ch := range subs {
				close(ch)
			}
		}
	}
}
