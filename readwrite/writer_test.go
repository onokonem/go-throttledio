package readwrite_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"testing"
	"time"

	"github.com/onokonem/go-throttledio/limiter"
	"github.com/onokonem/go-throttledio/readwrite"
)

const (
	testDuration = time.Second * 10
	interval     = time.Second * 1
	ticks        = 100
	concurency   = 10
	maxDeviation = 0.05
)

var testAmount = func() int64 {
	n := int64(1024 * 1024)
	for {
		startTime := time.Now()
		io.CopyN(ioutil.Discard, &noOpReader{}, n)
		if spent := time.Now().Sub(startTime); spent >= testDuration/10 {
			fmt.Printf("test amount: %d (%v)\n", n, spent)
			return n
		}
		n *= 2
	}
}()

func TestWriterDelay(t *testing.T) {
	l := limiter.NewController(interval, ticks, 0, 0).BornLimiter()
	cw := &countingWriter{w: ioutil.Discard}
	w := readwrite.NewWriter(cw, l, false)

	amount := testAmount / 2
	spent := cp(w, &noOpReader{}, amount)
	maxSpeed := int64(float64(cw.counter) / spent.Seconds())
	fmt.Printf("unthrotled speed: %d bytes in %v: %d bps\n", amount, spent, maxSpeed)

	if cw.counter != amount {
		t.Errorf("expected %d, got %d", testAmount, cw.counter)
	}

	cw.counter = 0
	l.SetCPS(maxSpeed / 2)

	amount = testAmount / 4
	spent = cp(w, &noOpReader{}, amount)
	speed := float64(maxSpeed / 2)
	realSpeed := float64(cw.counter) / spent.Seconds()
	deviation := (realSpeed - speed) / speed

	fmt.Printf("throtled speed: %2.2f bps, %d bytes in %v: %2.2f bps, deviation %3.3f\n", speed, amount, spent, realSpeed, deviation)

	if cw.counter != amount {
		t.Errorf("expected %d, got %d", testAmount, cw.counter)
	}
	if math.Abs(deviation) > maxDeviation {
		t.Errorf("deviation is too big: %f", deviation)
	}
}

func TestWriteDeadline(t *testing.T) {
	w := readwrite.NewWriter(ioutil.Discard, limiter.NewController(interval, ticks, 1, 1).BornLimiter(), false)

	timeout := interval * 2
	startTime := time.Now()
	w.SetDeadline(startTime.Add(timeout))

	_, err := w.Write(make([]byte, 1000))
	if err == nil || !errors.Is(err, readwrite.ErrDeadline) {
		t.Errorf("expected %v, got %v", readwrite.ErrDeadline, err)
	}

	spent := time.Now().Sub(startTime)
	if r := math.Abs(float64(spent-timeout)) / float64(timeout); r > 0.001 {
		t.Errorf("expected %v, got %v (%f)", timeout, spent, r)
	}
}

func TestWriteFragile(t *testing.T) {
	w := readwrite.NewWriter(ioutil.Discard, limiter.NewController(interval, ticks, 1, 1).BornLimiter(), true)

	_, err := w.Write(make([]byte, 1000))
	if err == nil || !errors.Is(err, readwrite.ErrExceeded) {
		t.Errorf("expected %v, got %v", readwrite.ErrExceeded, err)
	}
}

func TestWriteError(t *testing.T) {
	w := readwrite.NewWriter(&errWriter{}, limiter.NewController(interval, ticks, 0, 0).BornLimiter(), false)

	_, err := w.Write(make([]byte, 1000))
	if err == nil || !errors.Is(err, errWriterTest) {
		t.Errorf("expected %v, got %v", errWriterTest, err)
	}
}

func TestWriteEmpty(t *testing.T) {
	w := readwrite.NewWriter(ioutil.Discard, limiter.NewController(interval, ticks, 0, 0).BornLimiter(), false)

	n, err := w.Write(nil)
	if n != 0 || err != nil {
		t.Errorf("expected (0, nil), got (%d,%v)", n, err)
	}
}

type countingWriter struct {
	w       io.Writer
	counter int64
}

func (w *countingWriter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	w.counter += int64(n)
	return n, err
}

var errWriterTest = errors.New("test")

type errWriter struct{}

func (w *errWriter) Write(p []byte) (n int, err error) {
	return len(p) / 2, errWriterTest
}

var errPartialWrite = errors.New("partialWrite")

func cp(w io.Writer, r io.Reader, n int64) time.Duration {
	startTime := time.Now()

	cn, err := io.CopyN(w, r, n)
	if err != nil {
		panic(err)
	}

	if cn != n {
		panic(fmt.Errorf("expected %d, got %d: %w", n, cn, errPartialWrite))
	}

	return time.Now().Sub(startTime)
}

type noOpReader struct{}

func (r *noOpReader) Read(p []byte) (int, error) {
	return len(p), nil
}

func BenchmarkNoOp(b *testing.B) {
	io.CopyN(ioutil.Discard, &noOpReader{}, testAmount/100)
}

func BenchmarkWriter(b *testing.B) {
	io.CopyN(readwrite.NewWriter(ioutil.Discard, limiter.NewController(interval, ticks, 0, 0).BornLimiter(), true), &noOpReader{}, testAmount/100)
}
