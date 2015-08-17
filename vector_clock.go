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

func (v VectorClock) copy() VectorClock {
	clocks := make(map[int]uint64, len(v.clocks))
	for k, v := range v.clocks {
		clocks[k] = v
	}
	return VectorClock{name: v.name, clocks: clocks}
}

func (v VectorClock) Inc() VectorClock {
	newClock := v.copy()
	newClock.clocks[v.name] += 1
	return newClock
}

func (v VectorClock) Merge(o VectorClock) VectorClock {
	newClock := v.copy()
	for k, n := range o.clocks {
		if k == v.name && v.name != o.name { // merge external ticks only except it's the same type
			continue
		}
		if vv, exists := newClock.clocks[k]; !exists || n > vv {
			newClock.clocks[k] = n
		}
	}
	return newClock
}

func (v VectorClock) WithoutExternalTicks() VectorClock {
	newClock := NewVectorClock(v.name)
	newClock.clocks[v.name] = v.clocks[v.name]
	return newClock
}

func (v VectorClock) Get(name int) (uint64, bool) {
	vl, ok := v.clocks[name]
	return vl, ok
}

// Before compares the clock value of objects clock's own internal tick with given clock's external tick value.
// clock[v.name] < o.clock[v.name]
func (v VectorClock) Before(o VectorClock) bool {
	return v.compareTo(o) < 0
}

// After compares the clock value of objects clock's own tick with given clock's tick of same name.
// clock[v.name] > o.clock[v.name]
func (v VectorClock) After(o VectorClock) bool {
	return v.compareTo(o) > 0
}

func (v VectorClock) compareTo(o VectorClock) int {
	v1, _ := v.Get(v.name)
	o1, ok := o.Get(v.name)
	if !ok {
		o1 = 0
	}
	if v1 > o1 {
		return 1
	}
	if v1 < o1 {
		return -1
	}

	return 0
}

func (v VectorClock) IsEmpty() bool {
	return len(v.clocks) == 1 && v.clocks[v.name] == 0
}

// Equals compares internal and all external tick values.
func (v VectorClock) Equals(o VectorClock) bool {
	if len(v.clocks) != len(o.clocks) {
		return false
	}
	for k, v := range v.clocks {
		if o.clocks[k] != v {
			return false
		}
	}
	return true
}

func (v VectorClock) IsSameType(o VectorClock) bool {
	return v.name == o.name
}
