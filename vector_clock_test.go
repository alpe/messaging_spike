package main

import "testing"

func TestClocksMatchWikipediaExample(t *testing.T) {
	// given
	a := NewVCModel(A)
	b := NewVCModel(B)
	c := NewVCModel(C)

	// when
	if err := c.SendTo(b); err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if err := b.SendTo(a); err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if err := b.SendTo(c); err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if err := a.SendTo(b); err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if err := c.SendTo(a); err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if err := b.SendTo(c); err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	if err := c.SendTo(a); err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	// then
	if got, expected := a.Clock(A), 4; got != expected {
		t.Errorf("expected %d but got %d for A", expected, got)
	}
	if got, expected := a.Clock(B), 5; got != expected {
		t.Errorf("expected %d but got %d for B", expected, got)
	}
	if got, expected := a.Clock(C), 5; got != expected {
		t.Errorf("expected %d but got %d for C", expected, got)
	}
}
func TestAfter(t *testing.T) {
	if !NewVectorClock(A).Inc().After(NewVectorClock(A)) {
		t.Errorf("expected to be true")
	}
}

func TestWithMultipleConsumersOnlyOneShouldAct(t *testing.T) {
	// topic with A+B as consumers
	a := NewVCModel(A)
	otherA := NewVCModel(A)
	c := NewVCModel(C)

	if err := a.SendTo(c, "v1", "v2"); err != nil {
		t.Fatalf("%v", err)
	}
	if err := otherA.SendTo(c, "v3"); err == nil {
		t.Fatal("expected error")
	}
	expected := "v2"
	if c.state != expected {
		t.Errorf("expected state %q but was %q", expected, c.state)
	}
}

func TestWithDetached(t *testing.T) {
	// topic with A+B as consumers
	a := NewVCModel(A)
	detached := NewVCModel(B)
	c := NewVCModel(C)

	if err := a.SendTo(c, "v1", "v2"); err != nil {
		t.Fatalf("%v", err)
	}
	if err := detached.SendTo(c, "latest"); err != nil {
		t.Fatalf("%v", err)
	}

	expected := "latest"
	if c.state != expected {
		t.Errorf("expected state %q but was %q", expected, c.state)
	}
}
