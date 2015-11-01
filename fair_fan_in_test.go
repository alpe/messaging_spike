package messaging_spike

import (
	"fmt"
	"testing"
)

// any payload type

const NumberMessages = 10

// test non blocking round robin consumption
func TestFanInShouldMergeRoundRobin(t *testing.T) {
	names := []string{"a", "b", "c"}

	out := make(chan Payload, len(names)*NumberMessages)

	c := make([]chan Payload, len(names))
	for i, name := range names {
		c[i] = make(chan Payload, NumberMessages)
		produce(name, c[i])

	}

	// when
	go FanIn(c, out)
	// then
	for i := 0; i < NumberMessages; i++ {
		for _, name := range names {
			v := <-out
			if got, exp := v, newPayload(name, i); got != exp {
				t.Fatalf("expected %q but got %q", exp, got)
			}
		}
	}
}

func TestFanInShouldPruneClosedInputChannels(t *testing.T) {
	// given
	out := make(chan Payload, NumberMessages)

	in := make([]chan Payload, 2)
	in[0] = make(chan Payload)
	close(in[0])
	in[1] = make(chan Payload, NumberMessages)
	produce("active", in[1])
	// when
	go FanIn(in, out)
	// then
	for i := 0; i < NumberMessages; i++ {
		v := <-out
		if got, exp := v, newPayload("active", i); got != exp {
			t.Fatalf("expected %q but got %q", exp, got)
		}
	}
}
func TestFanInShouldCloseOutChannelWhenAllInputChannelsAreClosed(t *testing.T) {
	out := make(chan Payload)
	closedChannel := make(chan Payload)
	close(closedChannel)
	in := []chan Payload{closedChannel}
	// when
	FanIn(in, out)
	// then
	_, ok := <-out
	if got, exp := ok, false; got != exp {
		t.Errorf("expected %v but got %v", exp, got)
	}

}

func produce(name string, out chan<- Payload) {
	counter := 0
	for i := 0; i < NumberMessages; i++ {
		out <- newPayload(name, counter)
		counter += 1
	}
}

func newPayload(name string, counter int) Payload {
	return Payload(fmt.Sprintf("%v%d", name, counter))
}
