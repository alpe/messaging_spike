package main

import "testing"

func TestClocksMatchWikipediaExample(t *testing.T) {

	a := NewVCModel(A)
	b := NewVCModel(B)
	c := NewVCModel(C)

	c.SendTo(b)
	b.SendTo(a)

	b.SendTo(c)
	a.SendTo(b)

	c.SendTo(a)
	b.SendTo(c)
	c.SendTo(a)

	if got, expected := a.Clock(A), 4; got != expected {
		t.Errorf("expected '%s' but got '%s' for A", expected, got)
	}
	if got, expected := a.Clock(B), 5; got != expected {
		t.Errorf("expected '%s' but got '%s' for B", expected, got)
	}
	if got, expected := a.Clock(C), 5; got != expected {
		t.Errorf("expected '%s' but got '%s' for C", expected, got)
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
