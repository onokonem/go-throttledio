package atomic_test

import (
	"testing"
	"time"

	"github.com/onokonem/go-throttledio/internal/atomic"
)

func TestTime(t *testing.T) {
	expected := time.Time{}
	a := atomic.NewTime(expected)
	if !a.Get().IsZero() {
		t.Errorf("expected %#+v, got %#+v", expected, a)
	}

	expected = time.Now()
	a.Set(expected)
	if a.Get() != expected {
		t.Errorf("expected %#+v, got %#+v", expected, a)
	}

	expected = expected.Add(time.Hour)
	a.Set(expected)
	if a.Get() != expected {
		t.Errorf("expected %#+v, got %#+v", expected, a)
	}
}
