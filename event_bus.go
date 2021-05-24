package gopostgrespubsub

import (
	"errors"
	"sync"
)

var (
	ErrEventBusClosed  = errors.New("event bus closed")
	ErrChannelNotFound = errors.New("channel not found")
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

func (eb *EventBus) Subscribe(topic string) DataChannel {
	eb.rm.Lock()
	defer eb.rm.Unlock()

	ch := make(DataChannel, 1)
	if chans, found := eb.subscribers[topic]; found {
		eb.subscribers[topic] = append(chans, ch)
		return ch
	}

	eb.subscribers[topic] = []DataChannel{ch}
	return ch
}

func (eb *EventBus) Unsubscribe(topic string) error {
	//TODO: remover 'topic' do mapa de subscribers
	//?: tem que fechar alguma coisa nesse momento?
	return errors.New("not implemented")
}

func (eb *EventBus) Publish(topic string, data interface{}) error {
	eb.rm.RLock()
	defer eb.rm.RUnlock()

	if eb.closed {
		return ErrEventBusClosed
	}

	chans, found := eb.subscribers[topic]
	if !found {
		return ErrChannelNotFound
	}

	channels := make(DataChannelSlice, len(chans))
	copy(channels, chans)
	go func(data DataEvent, dataChans DataChannelSlice) {
		for _, ch := range dataChans {
			ch <- data
		}
	}(DataEvent{Data: data, Topic: topic}, channels)

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
