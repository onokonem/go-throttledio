package throttledio

import (
	"io"
	"sync/atomic"
	"time"

	"github.com/onokonem/go-throttledio/internal/counter"
	"github.com/onokonem/go-throttledio/internal/options"
	"golang.org/x/xerrors"
)

// default writer params
const (
	Interval = time.Second * 30
	Ticks    = uint(30)
	CPS      = float64(0)
	Timeout  = Interval * 2
)

// writer errors
var (
	ErrUnknownOption = xerrors.New("unknown option")
	ErrExceeded      = xerrors.New("bandwidth exceeded")
)

// Writer is a wrapper for io.Writer with throttling implemented
type Writer struct {
	writer  io.Writer
	counter *counter.Counter
	noerr   bool
	discard bool
	cps     atomic.Value
	sleep   time.Duration
	ticks   int
}

// NewWriter makes the Writer instance
func NewWriter(w io.Writer, opts ...WriterOption) *Writer {
	res := &Writer{
		writer:  w,
		noerr:   true,
		discard: false,
	}

	var (
		interval = Interval
		ticks    = Ticks
		cps      = CPS
	)

	for _, o := range opts {
		switch p := o.(type) {
		case *options.Interval:
			interval = p.V
		case *options.Ticks:
			ticks = p.V
		case *options.NoError:
			res.noerr = p.V
		case *options.Discard:
			res.discard = p.V
		case *options.Speed:
			cps = p.V
		default:
			panic(xerrors.Errorf("option %#+v: %w", o, ErrUnknownOption))
		}
	}

	res.counter = counter.NewCounter(interval, ticks)
	res.sleep = interval / time.Duration(ticks)
	res.ticks = int(ticks)
	res.SetSpeed(cps)

	return res
}

// SetSpeed is used to change throttling CPS on the fly
func (w *Writer) SetSpeed(cps float64) {
	w.cps.Store(cps)
	w.counter.Reset(cps)
}

// GetSpeed returns current CPS
func (w *Writer) GetSpeed() float64 {
	return w.cps.Load().(float64)
}

// Write method to implement io.Writer interface
func (w *Writer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	cps := w.GetSpeed()
	if cps <= 0 {
		n, err := w.writer.Write(p)
		w.counter.FillUp(int64(n))
		return n, err
	}

	if w.discard {
		return w.writeDiscard(p, cps)
	}

	return w.writeDelay(p, cps)
}

func (w *Writer) writeDiscard(p []byte, cps float64) (n int, err error) {
	b := p

	for len(b) > 0 {
		allowed := int(w.counter.FillUpToCap(int64(len(b)), cps))
		if allowed <= 0 {
			if w.noerr {
				return len(p), nil
			}
			return len(p) - len(b), ErrExceeded
		}

		n, err = w.writer.Write(b[:allowed])
		if n != allowed {
			w.counter.FillUp(int64(n - allowed))
		}
		if err != nil {
			return n, err
		}

		b = b[n:]
	}

	return len(p), nil
}

func (w *Writer) writeDelay(p []byte, cps float64) (n int, err error) {
	b := p

	for len(b) > 0 {
		allowed := int(w.counter.FillUpToCap(int64(len(b)), cps))
		if allowed <= 0 {
			time.Sleep(w.sleep)
			allowed = int(w.counter.FillUpToCap(int64(len(b)), cps))
		}

		n, err = w.writer.Write(b[:allowed])
		if n != allowed {
			w.counter.FillUp(int64(n - allowed))
		}
		if err != nil {
			return n, err
		}

		b = b[n:]
	}

	return len(p), nil
}
