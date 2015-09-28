package messaging_spike

import (
	"fmt"
)

// source events
type Coffee struct {
	customer string // globally unique for simplicity
	kind     string
}

type Sugar struct {
	customer string
	ordered  bool
}

const (
	takeAway = iota
	inHouse
)

type Place struct {
	customer string
	where    int
}

// state events

type Event interface{}

type OrderCreated struct {
	customer string // globally unique for simplicity
	kind     string
	sugar    bool
	where    int
}

func newOrderCreated(c Coffee, s Sugar, p Place) OrderCreated {
	return OrderCreated{
		customer: c.customer,
		kind:     c.kind,
		sugar:    s.ordered,
		where:    p.where,
	}
}

type PartialOrder struct {
	customer string
	coffee   *Coffee
	sugar    *Sugar
	place    *Place
}

// transient state object
type transientOrder struct {
	coffee *Coffee
	sugar  *Sugar
	place  *Place
}

func (t *transientOrder) merge(o PartialOrder) {
	// attributes are only updated once, therefore no overwrite/ merge conflict
	if o.coffee != nil {
		t.coffee = o.coffee
	}
	if o.sugar != nil {
		t.sugar = o.sugar
	}
	if o.place != nil {
		t.place = o.place
	}
}

func (t transientOrder) isComplete() bool {
	return t.coffee != nil && t.sugar != nil && t.place != nil
}

// event consumer
type CoffeeOrderConsumer struct {
	StateEvents    []Event // for simplicity: writing to StateEvents is persisting the event
	sugarStore     map[string]Sugar
	placeStore     map[string]Place
	ordersInFlight map[string]*transientOrder
}

func NewCoffeeOrderConsumer() *CoffeeOrderConsumer {
	return &CoffeeOrderConsumer{
		sugarStore:     make(map[string]Sugar),
		placeStore:     make(map[string]Place),
		ordersInFlight: make(map[string]*transientOrder),
		StateEvents:    make([]Event, 0),
	}
}

// process event
func (c *CoffeeOrderConsumer) OnEvent(e Event) error {
	switch ev := e.(type) {
	case Sugar:
		c.sugarStore[ev.customer] = ev
		return c.withInFlight(ev.customer, func(t *transientOrder) {
			t.sugar = &ev
		})
	case Place:
		c.placeStore[ev.customer] = ev
		return c.withInFlight(ev.customer, func(t *transientOrder) {
			t.place = &ev
		})
	case Coffee:
		s, sOk := c.sugarStore[ev.customer]
		p, pOk := c.placeStore[ev.customer]
		if sOk && pOk { // all dependencies are fulfilled, move on
			c.StateEvents = append(c.StateEvents, newOrderCreated(ev, s, p))
			return nil
		}
		// store/ persist state
		c.ordersInFlight[ev.customer] = &transientOrder{}
		return c.withInFlight(ev.customer, func(t *transientOrder) {
			t.coffee = &ev
		})
	default:
		return fmt.Errorf("unsupported event: %T", ev)
	}
}

func (c *CoffeeOrderConsumer) SourceEvents(events []Event) error {
	completedOrders := make([]OrderCreated, 0)
	for _, e := range events {
		switch ev := e.(type) {
		case Sugar:
			c.sugarStore[ev.customer] = ev
		case Place:
			c.placeStore[ev.customer] = ev
		case OrderCreated:
			completedOrders = append(completedOrders, ev)
		case PartialOrder:
			t, ok := c.ordersInFlight[ev.customer]
			if !ok {
				t = &transientOrder{}
				c.ordersInFlight[ev.customer] = t
			}
			t.merge(ev)
		default: // ignore
		}
	}
	// cleanup orders in flight due to random state events
	for _, e := range completedOrders {
		if _, ok := c.ordersInFlight[e.customer]; ok {
			delete(c.ordersInFlight, e.customer)
		}
	}
	return nil
}

func (c *CoffeeOrderConsumer) withInFlight(customer string, merge func(*transientOrder)) error {
	if i, ok := c.ordersInFlight[customer]; ok {
		merge(i)
		if i.isComplete() {
			c.StateEvents = append(c.StateEvents, newOrderCreated(*i.coffee, *i.sugar, *i.place))
			delete(c.ordersInFlight, customer)
			return nil
		}
		p := PartialOrder{customer: customer, coffee: i.coffee}
		if i.place != nil {
			p.place = i.place
		}
		if i.sugar != nil {
			p.sugar = i.sugar
		}
		c.StateEvents = append(c.StateEvents, p)
	}
	return nil
}
