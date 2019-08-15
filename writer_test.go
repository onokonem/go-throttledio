package throttledio_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/onokonem/go-throttledio"
	"golang.org/x/xerrors"
)

const (
	testDuration = time.Second * 10
	interval     = time.Second * 1
	ticks        = 100
	concurency   = 10
)

func TestWriterSpeed(t *testing.T) {
	const n = 1024 * 1024 * 100

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	cw := &countingWriter{w: ioutil.Discard}
	w := throttledio.NewWriter(
		cw,
		throttledio.SetDiscard(false),
		throttledio.SetNoError(true),
		throttledio.SetInterval(time.Second*5, 100),
		throttledio.SetSpeed(0),
	)

	spent := cp(w, r, n, 0)
	maxSpeed := float64(cw.counter) / spent.Seconds()
	fmt.Printf("unthrotled speed: %d bytes in %v: %2.2f bps\n", n, spent, maxSpeed)

	if cw.counter != n {
		t.Errorf("expected %d, got %d", n, cw.counter)
	}

	cw.counter = 0
	speed := maxSpeed / 2

	spent = cp(w, r, n, speed)

	realSpeed := float64(cw.counter) / spent.Seconds()
	deviation := (realSpeed - speed) / (speed / 100)

	fmt.Printf("throtled speed: %2.2f bps, %d bytes in %v: %2.2f bps, deviation %2.2f\n", speed, n, spent, realSpeed, deviation)

	if cw.counter != n {
		t.Errorf("expected %d, got %d", n, cw.counter)
	}
	if math.Abs(deviation) > 5 {
		t.Errorf("deviation is too big: %f", deviation)
	}
}

func TestWriterDiscard(t *testing.T) {
	const n = 1024 * 1024 * 100

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	cw := &countingWriter{w: ioutil.Discard}
	w := throttledio.NewWriter(
		cw,
		throttledio.SetDiscard(true),
		throttledio.SetNoError(true),
		throttledio.SetInterval(time.Second*5, 100),
		throttledio.SetSpeed(1),
	)

	cn, err := io.CopyN(w, r, n)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if cn != n {
		t.Errorf("expected %d, got %d", n, cn)
	}

	if cw.counter == n {
		t.Errorf("expected %d, got the same but should not", n)
	}

	w.SetSpeed(10000000000000)
	cw.counter = 0

	cn, err = io.CopyN(w, r, n)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if cn != n {
		t.Errorf("expected %d, got %d", n, cn)
	}

	//if cw.counter != n {
	//	t.Errorf("expected %d, got %d", n, cw.counter)
	//}
}

func TestWriterDiscardError(t *testing.T) {
	const n = 1024 * 1024 * 100

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	cw := &countingWriter{w: ioutil.Discard}
	w := throttledio.NewWriter(
		cw,
		throttledio.SetDiscard(true),
		throttledio.SetNoError(false),
		throttledio.SetInterval(time.Second*5, 100),
		throttledio.SetSpeed(1),
	)

	_, err := io.CopyN(w, r, n)
	if err == nil {
		t.Error("error was expected")
	}

	if !xerrors.Is(err, throttledio.ErrExceeded) {
		t.Error(err)
	}
}

func TestWriteEmpty(t *testing.T) {
	w := throttledio.NewWriter(ioutil.Discard)

	n, err := w.Write(nil)
	if n != 0 || err != nil {
		t.Errorf("expected (0, nil), got (%d,%v)", n, err)
	}
}

func TestFakeOption(t *testing.T) {
	err := testFakeOption()

	if err == nil {
		t.Error("error was expected")
	}

	if !xerrors.Is(err, throttledio.ErrUnknownOption) {
		t.Error(err)
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

func cp(w *throttledio.Writer, r io.Reader, n int64, speed float64) time.Duration {
	w.SetSpeed(speed)

	startTime := time.Now()

	cn, err := io.CopyN(w, r, n)
	if err != nil {
		panic(err)
	}

	if cn != n {
		panic(xerrors.Errorf("expected %d, got %d", n, cn))
	}

	return time.Now().Sub(startTime)
}

type fakeOption struct {
}

func (o *fakeOption) ItIsAWriterOption() {}
func (o *fakeOption) ItIsAReaderOption() {}

var (
	ErrUnexpected = xerrors.New("unexpected error")
)

func testFakeOption() (err error) {
	defer func() {
		er := recover()
		if er != nil {
			if ee, ok := er.(error); ok {
				err = ee
			} else {
				err = xerrors.Errorf("recovered: %#+v: %w", er, ErrUnexpected)
			}
		}
	}()
	_ = throttledio.NewWriter(nil, &fakeOption{})
	return xerrors.Errorf("there muste be panic, %w", ErrUnexpected)
}
