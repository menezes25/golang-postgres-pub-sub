package gopostgrespubsub

import (
	"errors"
	"sync"
)

var (
	ErrChannelNotFound = errors.New("channel not found")
)

type DataEvent struct {
	Data  interface{}
	Topic string
}

type DataChannel chan DataEvent

type DataChannelSlice []DataChannel

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string]DataChannelSlice),
	}
}

type EventBus struct {
	subscribers map[string]DataChannelSlice
	rm          sync.RWMutex
}

func (eb *EventBus) Subscribe(topic string, ch DataChannel) {
	eb.rm.Lock()
	defer eb.rm.Unlock()

	if chans, found := eb.subscribers[topic]; found {
		eb.subscribers[topic] = append(chans, ch)
		return
	}

	eb.subscribers[topic] = []DataChannel{ch}
}

func (eb *EventBus) Unsubscribe(topic string) error {
	//TODO: remover 'topic' do mapa de subscribers
	//?: tem que fechar alguma coisa nesse momento?
	return errors.New("not implemented")
}

func (eb *EventBus) Publish(topic string, data interface{}) error {
	eb.rm.RLock()
	defer eb.rm.RUnlock()
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
