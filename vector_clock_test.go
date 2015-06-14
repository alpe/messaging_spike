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
