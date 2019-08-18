package readwrite

import (
	"io"
	"time"

	"github.com/onokonem/go-throttledio/internal/atomic"
	"github.com/onokonem/go-throttledio/limiter"
)

const writeRetryDelay = time.Microsecond

// Writer is a wrapper for io.Writer with throttling implemented
type Writer struct {
	writer   io.Writer
	limiter  *limiter.Limiter
	fragile  bool
	deadline *atomic.Time
}

// NewWriter makes the Writer instance
// w in an underlaing io.Writer
// limiter is a limiter instance to be used to contron Writer bandwidth
// fragile flags controlss will the writer return an error on bandwidth exceeded,
// or will it retry until deadline.
func NewWriter(w io.Writer, limiter *limiter.Limiter, fragile bool) *Writer {
	return &Writer{
		writer:   w,
		limiter:  limiter,
		fragile:  fragile,
		deadline: atomic.NewTime(time.Time{}),
	}
}

// SetDeadline sets a deadline for the next Write.
func (w *Writer) SetDeadline(t time.Time) {
	w.deadline.Set(t)
}

// Write method to implement io.Writer interface
func (w *Writer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	b := p

	for len(b) > 0 {
		if deadline := w.deadline.Get(); !deadline.IsZero() && time.Now().After(deadline) {
			return len(p) - len(b), ErrDeadline
		}
		allowed := w.limiter.FillUp(int64(len(b)))
		if allowed <= 0 {
			if w.fragile {
				return len(p) - len(b), ErrExceeded
			}
			time.Sleep(writeRetryDelay)
			continue
		}
		n, err = w.writer.Write(b[:allowed])
		if left := int64(n) - allowed; left < 0 {
			w.limiter.FillUp(left)
		}
		if err != nil {
			return len(p) - len(b), err
		}

		b = b[n:]
	}

	return len(p), nil
}
