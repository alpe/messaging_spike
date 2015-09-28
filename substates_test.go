package messaging_spike

import (
	"math/rand"
	"reflect"
	"testing"
)

// Scenario:
// We receive events in almost random order. We act only if there is a specific event
// but data from other events needs to be available before.
// Constraints: we don't get events twice during processing. When we consume an event the previous
// one is marked as read.
func TestHandleAggregateEventBeforeReferencedEvents(t *testing.T) {
	// given
	events := []Event{
		Coffee{"Alex", "flat white"}, // requires Sugar and Place
		Sugar{"Alex", false},
		Place{"Alex", takeAway},
		Sugar{"Piotr", false},
		Place{"Piotr", inHouse},
		Coffee{"Piotr", "Americano"},
	}

	// when
	o := NewCoffeeOrderConsumer()
	for _, e := range events {
		if err := o.OnEvent(e); err != nil {
			t.Fatalf("unexpected error %s", err)
		}
	}

	// then
	s := o.StateEvents
	if got, exp := len(s), 4; got != exp {
		t.Fatalf("expected %d but got %d", exp, got)
	}
	got := s[2].(OrderCreated)
	exp := OrderCreated{
		customer: "Alex", kind: "flat white", sugar: false, where: takeAway,
	}
	if !reflect.DeepEqual(got, exp) {
		t.Errorf("excpect %#v but got %#v", exp, got)
	}
}

func TestSourceRandomOrderOfEvents(t *testing.T) {
	for i := 0; i < 100; i++ { // run test multiple times
		// given
		events := []Event{
			Coffee{"Alex", "Flat white"},
			Sugar{"Alex", false},
			Place{"Alex", takeAway},
			Coffee{"Bob", "Americano"},
			Sugar{"Bob", true},
			Place{"Cesar", inHouse},
			Coffee{"Cesar", "Cappucino"},
			Sugar{"Daniel", false},
			Place{"Emil", takeAway},
			Coffee{"Fred", "Black coffee"},
			Sugar{"Fred", false},
			Place{"Fred", takeAway},
		}

		p := NewCoffeeOrderConsumer()
		for _, e := range events {
			if err := p.OnEvent(e); err != nil {
				t.Fatalf("unexpected error %s", err)
			}
		}
		s := NewCoffeeOrderConsumer()
		random := shuffle(append(events, p.StateEvents...)...)

		// when
		if err := s.SourceEvents(random); err != nil {
			t.Fatalf("unexpected error %s", err)
		}

		// then
		p.StateEvents = []Event{}
		if got, exp := s, p; !reflect.DeepEqual(got, exp) {
			t.Errorf("expected %#v but got %#v", exp, got)
		}
	}
}

func shuffle(a ...Event) []Event {
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
	return a
}
