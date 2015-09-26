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
	newState    string // for simplicity, state value is unique
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
	StateEvents              []ClockedEvent // for simplicity: writing to StateEvents is persisting the event
	beforeProcessingCallback ClockedEventCallback
	eventLog                 *failOnDuplicatesEventLog
	autoStartProcessing      bool
}

func NewAutoStartConsumer(name int) *SourceProcessConsumer {
	c := NewManualStartConsumer(name)
	c.autoStartProcessing = true
	return c
}

func NewManualStartConsumer(name int) *SourceProcessConsumer {
	return &SourceProcessConsumer{
		name:         name,
		vectorClock:  NewVectorClock(name),
		sourcedClock: NewVectorClock(name),
		StateEvents:  make([]ClockedEvent, 0),
		Mode:         ModeSourcing,
		beforeProcessingCallback: func(ClockedEvent) {},
		eventLog:                 newFailOnDuplicatesEventLog(),
	}
}

func (c *SourceProcessConsumer) OnEvent(e ClockedEvent) error {
	if c.autoStartProcessing && c.Mode == ModeSourcing && c.IsSourcingCompleted(e) {
		c.enableProcessingMode()
	}
	if c.Mode == ModeProcessing {
		return c.processEvent(e)
	}
	return c.sourceEvent(e)
}

func (c *SourceProcessConsumer) IsSourcingCompleted(e ClockedEvent) bool {
	if c.Mode == ModeProcessing {
		return true
	}
	if _, ok := e.(*InternalConsumerUpdatedMessage); ok { // we source our internal state messages first by convention
		return false
	}
	return c.clocksSynced()
}

func (c *SourceProcessConsumer) enableProcessingMode() {
	fmt.Printf("switching to processing mode: %+v\n", c)
	c.Mode = ModeProcessing
}

func (c *SourceProcessConsumer) DoSourcing() {
	c.sourcedClock = c.vectorClock
	c.Mode = ModeSourcing
	fmt.Printf("switching to sourcing mode: %+v\n", c)
}

func (c *SourceProcessConsumer) clocksSynced() bool {
	return c.sourcedClock.Equals(c.vectorClock)
}
func (c *SourceProcessConsumer) DoProcessing() error {
	if !c.clocksSynced() {
		return fmt.Errorf("internal clocks out of sync")
	}
	c.enableProcessingMode()
	return nil
}

func (c *SourceProcessConsumer) sourceEvent(e ClockedEvent) error {
	fmt.Printf("sourcing event: %#v\n", e)
	if !e.VectorClock().After(c.sourcedClock) {
		fmt.Printf("skipping message. already ahead: clock at  %+v; event: %#v\n", c.sourcedClock, e)
		return nil
	}

	c.sourcedClock = c.sourcedClock.Merge(e.VectorClock().WithoutExternalTicks())
	switch ev := e.(type) {
	case *ExternalEventMessage:
		if err := c.eventLog.Add(ev); err != nil {
			return err
		}
	case *InternalConsumerUpdatedMessage:
		c.state = ev.newState
		c.vectorClock = e.VectorClock()
	default:
		return fmt.Errorf("can not handle %T", e)
	}

	// check sourcing completed
	if c.autoStartProcessing && c.IsSourcingCompleted(e) {
		c.enableProcessingMode()
	}
	return nil
}

func (c *SourceProcessConsumer) processEvent(e ClockedEvent) error {
	c.beforeProcessingCallback(e)
	fmt.Printf("processing event: %#v\n", e)

	if !e.VectorClock().After(c.vectorClock) {
		return fmt.Errorf("recieved message out or order: %+v, my %+v: %+v\n", e.VectorClock(), c.vectorClock, e)
	}

	c.vectorClock = c.vectorClock.Inc().Merge(e.VectorClock())
	switch ev := e.(type) {
	case *ExternalEventMessage:
		if err := c.eventLog.Add(ev); err != nil {
			return err
		}
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

type failOnDuplicatesEventLog struct {
	msg map[string]struct{}
}

func newFailOnDuplicatesEventLog() *failOnDuplicatesEventLog {
	return &failOnDuplicatesEventLog{msg: make(map[string]struct{})}
}

func (f *failOnDuplicatesEventLog) Add(e *ExternalEventMessage) error {
	if _, ok := f.msg[e.newState]; ok {
		return fmt.Errorf("already exists: %q", e.newState)
	}
	f.msg[e.newState] = struct{}{}
	return nil

}
