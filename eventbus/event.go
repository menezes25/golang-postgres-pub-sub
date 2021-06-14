package eventbus

import "sync"

type Topic struct {
	Type    string      `json:"type,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

func Merge(cs ...<-chan Topic) <-chan Topic {
	// https://blog.golang.org/pipelines#TOC_4.
	var wg sync.WaitGroup
	out := make(chan Topic, 1)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan Topic) {
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
