package main

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

func (v VectorClock) Get(name int) (uint64, bool) {
	vl, ok := v.clocks[name]
	return vl, ok
}

type XXXMessage struct {
	vectorClock VectorClock
	newState    string
}

type VCModel struct {
	vectorClock VectorClock
	state       string
	name        int
}

func (f *VCModel) Clock(name int) int {
	v, _ := f.vectorClock.Get(name)
	return int(v)
}

func (f *VCModel) Receive(e XXXMessage) error {
	f.vectorClock = f.vectorClock.Inc()
	f.vectorClock.Merge(e.vectorClock)
	f.state = e.newState
	return nil
}

func (f *VCModel) SendTo(v *VCModel) *VCModel {
	f.vectorClock = f.vectorClock.Inc()
	v.Receive(XXXMessage{vectorClock: f.vectorClock})
	return f
}

func NewVCModel(name int) *VCModel {
	return &VCModel{
		name:        name,
		vectorClock: NewVectorClock(name),
	}
}
