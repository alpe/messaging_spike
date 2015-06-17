package main

import "fmt"

const (
	A = iota
	B
	C
)

// Immutable value type
type VectorClock struct {
	clocks map[int]uint64
	name   int
}

func NewVectorClock(name int) VectorClock {
	return VectorClock{
		name:   name,
		clocks: map[int]uint64{name: 0},
	}
}

func (v VectorClock) Inc() VectorClock {
	v.clocks[v.name] += 1
	return v
}
func (v VectorClock) Merge(o VectorClock) VectorClock {
	for k, n := range o.clocks {
		if k == v.name {
			continue
		}
		if vv, exists := v.clocks[k]; !exists || n > vv {
			v.clocks[k] = n
		}
	}
	return v
}

func (v VectorClock) Get(name int) uint64 {
	return v.clocks[name]
}
func (v VectorClock) IsBefore(o VectorClock) bool {
	return v.clocks[v.name] < o.Get(v.name)
}

// messages
type XXXMessage struct {
	vectorClock VectorClock
	newState    string
}

// model

type Receiver interface {
	Receive(XXXMessage) error
}

type VCModel struct {
	vectorClock VectorClock
	state       string
	name        int
}

func NewVCModel(name int) *VCModel {
	return &VCModel{
		name:        name,
		vectorClock: NewVectorClock(name),
	}
}

func (f *VCModel) Clock(name int) int {
	v := f.vectorClock.Get(name)
	return int(v)
}

func (f *VCModel) Receive(msg XXXMessage) error {
	fmt.Printf("Receiving %+v", msg)
	if msg.vectorClock.IsBefore(f.vectorClock) {
		return fmt.Errorf("state is ahead")
	}
	f.vectorClock = f.vectorClock.Inc()
	f.vectorClock.Merge(msg.vectorClock)
	f.state = msg.newState
	return nil
}

// fail and return the first error
func (f *VCModel) SendTo(r Receiver, msgs ...string) error {
	if len(msgs) == 0 {
		return f.SendTo(r, "")
	}
	for _, m := range msgs {
		if err := f.sendTo(r, m); err != nil {
			return err
		}
	}
	return nil
}

func (f *VCModel) sendTo(r Receiver, msg string) error {
	f.vectorClock = f.vectorClock.Inc()
	return r.Receive(XXXMessage{vectorClock: f.vectorClock, newState: msg})
}

func (f *VCModel) Reset() {
	f.vectorClock = NewVectorClock(f.name)
	f.state = ""
}
