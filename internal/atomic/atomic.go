package atomic

import (
	"sync/atomic"
	"time"
)

// Time is to store time.Time atomically
type Time atomic.Value

// Get returns a value stored
func (a *Time) Get() time.Time {
	return (*atomic.Value)(a).Load().(time.Time)
}

// Set stores a provided value
func (a *Time) Set(t time.Time) {
	(*atomic.Value)(a).Store(t)
}

// NewTime creates a new Time instabce and set it to a provided value
func NewTime(t time.Time) *Time {
	a := new(Time)
	(*atomic.Value)(a).Store(t)
	return a
}
