package gopostgrespubsub

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var eb = NewEventBus()

func TestEventBus(t *testing.T) {
	ch1 := make(chan DataEvent, 1)
	ch2 := make(chan DataEvent, 1)
	ch3 := make(chan DataEvent, 1)

	eb.Subscribe("t1", ch1)
	eb.Subscribe("t2", ch2)
	eb.Subscribe("t3", ch3)

	go publishTo("t1", "huaaaaa 1")
	go publishTo("t2", "hueeeee 2")

	for {
		select {
		case d := <-ch1:
			go printDataEvent("ch1", d)
		case d := <-ch2:
			go printDataEvent("ch2", d)
		case d := <-ch3:
			go printDataEvent("ch3", d)
		}
	}
}

func publishTo(topic string, data string) {
	for {
		eb.Publish(topic, data)
		<-time.After(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}
}

func printDataEvent(ch string, data DataEvent) {
	fmt.Printf("channel: %s | topic: %s | data_event: %v\n", ch, data.Topic, data.Data)
}
