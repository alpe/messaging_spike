package messaging_spike

import (
	"errors"
)

var ErrOptimisticLock = errors.New("optimistic lock. versions out of sync")

// events

type ModelEvent struct {
	seqID        uint16
	newState     string
	modelVersion uint16
}

type ErrorEvent struct {
	err error
	msg ModelEvent
}

// FooModel is a random persistent model, which would be an aggregate in DDD.
type FooModel struct {
	version uint16
	state   string
}

func (f *FooModel) OnEvent(e ModelEvent) error {
	if f.version != e.modelVersion {
		return ErrOptimisticLock
	}
	f.version++
	f.state = e.newState
	return nil
}

// ModelEventConsumer is a simple consumer in EIPatterns.
// Although Listen() has to be triggered by the tests.
type ModelEventConsumer struct {
	in           <-chan ModelEvent
	errChan      chan<- ErrorEvent
	rollingModel *FooModel
}

func NewModelEventConsumer(in <-chan ModelEvent, errChan chan<- ErrorEvent) *ModelEventConsumer {
	c := &ModelEventConsumer{in: in, errChan: errChan, rollingModel: &FooModel{version: 0, state: "init"}}
	return c
}

func (c *ModelEventConsumer) Listen() {
	e, ok := <-c.in
	if !ok {
		return
	}
	if err := c.rollingModel.OnEvent(e); err != nil {
		c.errChan <- ErrorEvent{err, e}
	}
}

func (c ModelEventConsumer) CurrentModel() FooModel {
	return *c.rollingModel
}
