package main

import (
	"reflect"
	"testing"
)

func TestConsumerShouldHandleModelUpdate(t *testing.T) {
	// given
	topic := make(chan ModelEvent, 1)
	c := NewModelEventConsumer(topic, nil)
	m := c.CurrentModel()

	// when
	topic <- ModelEvent{1, "myFirstState", m.version}
	c.Listen()

	// then
	var expectedVersion uint16 = 1
	if got := c.CurrentModel().version; got != expectedVersion {
		t.Errorf("model version should be %v but was %v", expectedVersion, got)
	}
	var expectedState = "myFirstState"
	if got := c.CurrentModel().state; got != expectedState {
		t.Errorf("model state should be %v but was %v", expectedState, got)
	}
}

func TestConsumerShouldHandleResubmissions(t *testing.T) {
	// given
	topic := make(chan ModelEvent, 1)
	errChan := make(chan ErrorEvent, 1)
	c := NewModelEventConsumer(topic, errChan)
	m := c.CurrentModel()

	// when
	duplicateEvent := ModelEvent{1, "myFirstState", m.version}
	topic <- duplicateEvent
	c.Listen()
	topic <- duplicateEvent
	c.Listen()
	close(errChan)

	// then
	expected := FooModel{version: 1, state: "myFirstState"}
	if got := c.CurrentModel(); !reflect.DeepEqual(got, expected) {
		t.Errorf("model should be %v but was %v", expected, got)
	}
	// and error was received
	if e, ok := <-errChan; !ok {
		t.Errorf("expected error msg")
	} else {
		expected := ErrorEvent{ErrOptimisticLock, duplicateEvent}
		if got := e; !reflect.DeepEqual(got, expected) {
			t.Errorf("event should be %+v but was %+v", expected, got)
		}
	}
}
