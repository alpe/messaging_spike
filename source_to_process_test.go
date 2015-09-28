package messaging_spike

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestSourceToProcessWithEventsOrdered(t *testing.T) {
	// given
	for i, spec := range []struct {
		q     *MessageQueue
		order int
	}{
		{NewIsolatedClocksMessageQueue(B, C), ByTimeLine},
		{NewSharedClocksMessageQueue(B, C), ByTimeLine},
		{NewIsolatedClocksMessageQueue(B, C), ByProducer},
		{NewSharedClocksMessageQueue(B, C), ByProducer},
	} {
		processingConsumer := NewAutoStartConsumer(A)

		q := spec.q.Add(B, "b1").Add(C, "c1").Add(B, "b2").Add(C, "c2")

		for _, e := range q.EventStream(ByTimeLine) {
			if err := processingConsumer.OnEvent(e); err != nil {
				t.Fatalf("queue: %d: unexpected error %s", i, err)
			}
		}
		if got, exp := len(processingConsumer.StateEvents), 4; got != exp {
			t.Fatalf("queue: %d: expected %d but got %d", i, exp, got)
		}

		// when sourcing and consumes previous state events and source events
		sourcingConsumer := NewAutoStartConsumer(A)
		sourcingConsumer.beforeProcessingCallback = func(e ClockedEvent) {
			t.Fatalf("queue: %d: there should be nothing to process", i)
		}
		for _, e := range append(processingConsumer.StateEvents, q.EventStream(spec.order)...) {
			if err := sourcingConsumer.OnEvent(e); err != nil {
				t.Fatalf("queue: %d: unexpected error %s", i, err)
			}
		}
		// then no new state event was written
		if got, exp := len(sourcingConsumer.StateEvents), 0; got != exp {
			t.Errorf("queue: %d: expected %d but got %d", i, exp, got)
		}
		// then has switched to processing mode
		if got, exp := sourcingConsumer.Mode, ModeProcessing; got != exp {
			t.Errorf("queue: %d: expected %d but got %d", i, exp, got)
		}
	}
}

func TestSwitchBackProcessToSourcing(t *testing.T) {
	// given two consumers
	q := NewSharedClocksMessageQueue(B, C).Add(B, "b1").Add(B, "b2").Add(C, "c1").Add(B, "b3").Add(B, "b4").Add(C, "c2")
	consumers := []*SourceProcessConsumer{NewManualStartConsumer(A), NewManualStartConsumer(A)}

	for i, e := range q.EventStream(ByTimeLine) {
		c1, c2 := consumers[i%2], consumers[(i+1)%2]
		// when one consumer processes
		if err := c1.DoProcessing(); err != nil {
			t.Fatalf("unexpected error %s", err)
		}
		if err := c1.OnEvent(e); err != nil {
			t.Fatalf("queue: %d: unexpected error %s", i, err)
		}
		// and one consumer sources
		c2.DoSourcing()
		if err := c2.OnEvent(c1.StateEvents[0]); err != nil {
			t.Fatalf("unexpected error %s", err)
		}

		if err := c2.OnEvent(e); err != nil {
			t.Fatalf("unexpected error %s", err)
		}

		c1.StateEvents = make([]ClockedEvent, 0)

		// then sourcing should be completed
		if ok := c2.IsSourcingCompleted(e); !ok {
			t.Fatal("sourcing not completed")
		}

		// and internal state should be the same
		if got, exp := c2.state, c1.state; got != exp {
			t.Fatalf("expected %v but got %v", exp, got)
		}
		if got, exp := c2.vectorClock, c1.vectorClock; !reflect.DeepEqual(got, exp) {
			t.Fatalf("expected %v but got %v", exp, got)
		}
		if got, exp := c2.StateEvents, c1.StateEvents; !reflect.DeepEqual(got, exp) {
			t.Fatalf("expected %v but got %v", exp, got)
		}
	}
	for _, c := range consumers {
		if got, exp := c.state, "c2"; got != exp {
			t.Errorf("expected %q but got %q", exp, got)
		}
	}

}

const (
	ByTimeLine = iota
	ByProducer
)

type MessageQueue struct {
	vc             map[int]VectorClock
	timeLine       []ClockedEvent
	incrementClock func(int) VectorClock
}

func NewIsolatedClocksMessageQueue(producers ...int) *MessageQueue {
	m := make(map[int]VectorClock, len(producers))
	for _, p := range producers {
		m[p] = NewVectorClock(p)
	}
	return &MessageQueue{
		vc:       m,
		timeLine: make([]ClockedEvent, 0),
		incrementClock: func(producer int) VectorClock {
			return m[producer].Inc()
		},
	}
}

func NewSharedClocksMessageQueue(producers ...int) *MessageQueue {
	m := make(map[int]VectorClock, len(producers))
	for _, p := range producers {
		m[p] = NewVectorClock(p)
	}
	return &MessageQueue{
		vc:       m,
		timeLine: make([]ClockedEvent, 0),
		incrementClock: func(producer int) VectorClock {
			c := m[producer].Inc()
			for _, v := range m {
				c = c.Merge(v)
			}
			return c
		},
	}
}

func (q MessageQueue) New() *MessageQueue {
	return &MessageQueue{
		vc:             q.vc,
		timeLine:       make([]ClockedEvent, 0),
		incrementClock: q.incrementClock,
	}
}

func (q *MessageQueue) Add(producer int, newState string) *MessageQueue {
	var newClock VectorClock
	for i, x := 0, rand.Int()%1000; i < x; i++ { // random clock step
		newClock = q.incrementClock(producer)
		q.vc[producer] = newClock
	}
	q.timeLine = append(q.timeLine, &ExternalEventMessage{vectorClock: newClock, newState: newState})
	return q
}
func (q *MessageQueue) EventStream(order int) []ClockedEvent {
	switch order {
	case ByProducer:
		pGroups := make(map[int][]ClockedEvent)
		for _, v := range q.timeLine {
			producer := v.VectorClock().name
			pGroups[producer] = append(pGroups[producer], v)
		}
		e := make([]ClockedEvent, 0)
		for _, v := range pGroups {
			e = append(e, v...)
		}
		return e
	default:
		return q.timeLine
	}
}
