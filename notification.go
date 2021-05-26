package gopostgrespubsub

import "sync"

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

func Merge(cs ...<-chan Event) <-chan Event {
	// https://blog.golang.org/pipelines#TOC_4.
	var wg sync.WaitGroup
	out := make(chan Event, 1)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan Event) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
