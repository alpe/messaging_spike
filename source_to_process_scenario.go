package main

import (
	"fmt"
)

const (
	ModeSourcing = iota
	ModeProcessing
)

type ClockedEvent interface {
	VectorClock() VectorClock
}

// Any events created by some Event Producer
type ExternalEventMessage struct {
	vectorClock VectorClock
	newState    string
}

func (e ExternalEventMessage) VectorClock() VectorClock {
	return e.vectorClock
}

// state events created by SourceProcessConsumer
type InternalConsumerUpdatedMessage struct {
	vectorClock VectorClock
	newState    string
}

func (e InternalConsumerUpdatedMessage) VectorClock() VectorClock {
	return e.vectorClock
}

type ClockedEventCallback func(ClockedEvent)

type SourceProcessConsumer struct {
	vectorClock              VectorClock
	sourcedClock             VectorClock
	state                    string
	name                     int
	Mode                     int
	StateEvents              []ClockedEvent
	beforeProcessingCallback ClockedEventCallback
}

func NewAutoSwitchConsumer(name int) *SourceProcessConsumer {
	return &SourceProcessConsumer{
		name:         name,
		vectorClock:  NewVectorClock(name),
		sourcedClock: NewVectorClock(name),
		StateEvents:  make([]ClockedEvent, 0), // for simplicity: writing to StateEvents is persisting the event
		Mode:         ModeSourcing,
		beforeProcessingCallback: func(ClockedEvent) {},
	}
}

func (c *SourceProcessConsumer) onEvent(e ClockedEvent) error {
	if c.isSourcingCompleted(e) {
		return c.processEvent(e)
	}
	return c.sourceEvent(e)
}

func (c *SourceProcessConsumer) isSourcingCompleted(e ClockedEvent) bool {
	if c.Mode == ModeProcessing {
		return true
	}

	if _, ok := e.(*InternalConsumerUpdatedMessage); ok { // we source our internal state messages first by convention
		return false
	}
	if c.sourcedClock.Equals(c.vectorClock) {
		c.enableProcessingMode()
		return true
	}
	return false
}
func (c *SourceProcessConsumer) enableProcessingMode() {
	fmt.Printf("switching to processing mode: %+v", c)
	c.Mode = ModeProcessing
}

func (c *SourceProcessConsumer) sourceEvent(e ClockedEvent) error {
	fmt.Printf("sourcing event: %#v\n", e)
	c.sourcedClock = c.sourcedClock.Merge(e.VectorClock().WithoutExternalTicks())
	switch ev := e.(type) {
	case *ExternalEventMessage:
	case *InternalConsumerUpdatedMessage:
		c.state = ev.newState
		c.vectorClock = e.VectorClock()
	default:
		return fmt.Errorf("can not handle %T", e)
	}

	// check sourcing completed
	if c.sourcedClock.Equals(c.vectorClock) {
		c.enableProcessingMode()
	}

	return nil
}

func (c *SourceProcessConsumer) processEvent(e ClockedEvent) error {
	c.beforeProcessingCallback(e)
	fmt.Printf("processing event: %#v\n", e)

	if !e.VectorClock().After(c.vectorClock) {
		return fmt.Errorf("recieved message out or order: %+v, my %+v\n", e.VectorClock(), c.vectorClock)
	}

	c.vectorClock = c.vectorClock.Inc().Merge(e.VectorClock())
	switch ev := e.(type) {
	case *ExternalEventMessage:
		c.state = ev.newState
		c.storeUpdate()
	default:
		return fmt.Errorf("can not handle %T", e)
	}
	return nil
}
func (c *SourceProcessConsumer) storeUpdate() {
	c.vectorClock = c.vectorClock.Inc()
	c.StateEvents = append(c.StateEvents, &InternalConsumerUpdatedMessage{
		vectorClock: c.vectorClock,
		newState:    c.state,
	})
}
