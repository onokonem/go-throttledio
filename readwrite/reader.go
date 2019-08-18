package readwrite

import (
	"io"
	"time"

	"github.com/onokonem/go-throttledio/internal/atomic"
	"github.com/onokonem/go-throttledio/limiter"
)

const readRetryDelay = time.Microsecond

// Reader is a wrapper for io.Reader with throttling implemented
type Reader struct {
	r        io.Reader
	limiter  *limiter.Limiter
	fragile  bool
	deadline *atomic.Time
}

// NewReader makes the Reader instance
// r in an underlaing io.Reader
// limiter is a limiter instance to be used to contron Reader bandwidth
// fragile flags controls will the reader return an error on bandwidth exceeded,
// or will it retry until deadline.
func NewReader(r io.Reader, limiter *limiter.Limiter, fragile bool) *Reader {
	return &Reader{
		r:        r,
		limiter:  limiter,
		fragile:  fragile,
		deadline: atomic.NewTime(time.Time{}),
	}
}

// SetDeadline sets a deadline for the next Read.
func (r *Reader) SetDeadline(t time.Time) {
	r.deadline.Set(t)
}

// Read method to implement io.Reader interface
func (r *Reader) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	for {
		if deadline := r.deadline.Get(); !deadline.IsZero() && time.Now().After(deadline) {
			return 0, ErrDeadline
		}
		allowed := r.limiter.FillUp(int64(len(p)))
		if allowed <= 0 {
			if r.fragile {
				return 0, ErrExceeded
			}
			time.Sleep(readRetryDelay)
			continue
		}

		n, err = r.r.Read(p[:allowed])
		if left := int64(n) - allowed; left < 0 {
			r.limiter.FillUp(left)
		}

		return n, err
	}
}
