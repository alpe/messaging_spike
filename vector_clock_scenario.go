package main

import "fmt"

type VCMessage struct {
	vectorClock VectorClock
	newState    string
}

// model

type Receiver interface {
	Receive(VCMessage) error
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
	v, _ := f.vectorClock.Get(name)
	return int(v)
}

func (f *VCModel) Receive(msg VCMessage) error {
	fmt.Printf("Receiving %+v", msg)
	if msg.vectorClock.Before(f.vectorClock) {
		return fmt.Errorf("state is ahead")
	}
	f.vectorClock = f.vectorClock.Inc().Merge(msg.vectorClock)
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
	return r.Receive(VCMessage{vectorClock: f.vectorClock, newState: msg})
}

func (f *VCModel) Reset() {
	f.vectorClock = NewVectorClock(f.name)
	f.state = ""
}
