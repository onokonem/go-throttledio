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

func TestReaderDelay(t *testing.T) {
	l := limiter.NewController(interval, ticks, 0, 0).BornLimiter()
	cw := &countingWriter{w: ioutil.Discard}
	r := readwrite.NewReader(&noOpReader{}, l, false)

	amount := testAmount / 2
	spent := cp(cw, r, amount)
	maxSpeed := int64(float64(cw.counter) / spent.Seconds())
	fmt.Printf("unthrotled speed: %d bytes in %v: %d bps\n", amount, spent, maxSpeed)

	if cw.counter != amount {
		t.Errorf("expected %d, got %d", testAmount, cw.counter)
	}

	cw.counter = 0
	l.SetCPS(maxSpeed / 2)

	amount = testAmount / 4
	spent = cp(cw, r, amount)
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

func TestReadDeadline(t *testing.T) {
	r := readwrite.NewReader(&noOpReader{}, limiter.NewController(interval, ticks, 1, 1).BornLimiter(), false)

	timeout := interval * 2
	startTime := time.Now()
	r.SetDeadline(startTime.Add(timeout))

	_, err := io.CopyN(ioutil.Discard, r, 1000)
	if err == nil || !errors.Is(err, readwrite.ErrDeadline) {
		t.Errorf("expected %v, got %v", readwrite.ErrDeadline, err)
	}

	spent := time.Now().Sub(startTime)
	if r := math.Abs(float64(spent-timeout)) / float64(timeout); r > 0.001 {
		t.Errorf("expected %v, got %v (%f)", timeout, spent, r)
	}
}

func TestReadFragile(t *testing.T) {
	r := readwrite.NewReader(&noOpReader{}, limiter.NewController(interval, ticks, 1, 1).BornLimiter(), true)

	_, err := io.CopyN(ioutil.Discard, r, 1000)
	if err == nil || !errors.Is(err, readwrite.ErrExceeded) {
		t.Errorf("expected %v, got %v", readwrite.ErrExceeded, err)
	}
}

func TestReadError(t *testing.T) {
	r := readwrite.NewReader(&errReader{}, limiter.NewController(interval, ticks, 0, 0).BornLimiter(), false)

	_, err := r.Read(make([]byte, 1000))
	if err == nil || !errors.Is(err, errReadTest) {
		t.Errorf("expected %v, got %v", errReadTest, err)
	}
}

func TestReadEmpty(t *testing.T) {
	r := readwrite.NewReader(&noOpReader{}, limiter.NewController(interval, ticks, 0, 0).BornLimiter(), false)

	n, err := r.Read(nil)
	if n != 0 || err != nil {
		t.Errorf("expected (0, nil), got (%d,%v)", n, err)
	}
}

var errReadTest = errors.New("test")

type errReader struct{}

func (r *errReader) Read(p []byte) (n int, err error) {
	return len(p) / 2, errReadTest
}

func BenchmarkReader(b *testing.B) {
	io.CopyN(ioutil.Discard, readwrite.NewReader(&noOpReader{}, limiter.NewController(interval, ticks, 0, 0).BornLimiter(), true), testAmount/100)
}
