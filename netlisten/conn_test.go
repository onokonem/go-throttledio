package netlisten_test

import (
	"fmt"
	"io"
	"math"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/onokonem/go-throttledio/limiter"
	"github.com/onokonem/go-throttledio/netlisten"
	"github.com/onokonem/go-throttledio/readwrite"
	"golang.org/x/xerrors"
)

func TestSetWriteCPS(t *testing.T) {
	interval := time.Second * 3

	l := listen()

	// just to increase the cover
	l.(*netlisten.Listener).WriteLimiter()

	reqSpeed := int64(100000)

	go func() {
		d := netlisten.NewDialer(
			net.Dialer{},
			limiter.NewController(Interval, Ticks, 0, 0),
			limiter.NewController(Interval, Ticks, 0, 0),
		)
		conn, err := d.Dial(l.Addr().Network(), l.Addr().String())
		if err != nil {
			panic(err)
		}
		conn.(*netlisten.Conn).SetWriteCPS(reqSpeed)
		io.Copy(conn, &noOpReader{})
	}()

	server := accept(l, new(int64))
	go server.read()

	startTime := time.Now()
	time.Sleep(interval * 5 / 2)

	curTime := time.Now()
	speed := float64(atomic.LoadInt64(server.total)) / curTime.Sub(startTime).Seconds()

	deviation := (float64(reqSpeed) - speed) / float64(reqSpeed)

	fmt.Printf("Limit: %d: speed: %f (d: %3.3f)\n", reqSpeed, speed, deviation)

	if math.Abs(deviation) > MaxDeviation {
		t.Errorf("deviation too big: %3.3f", deviation)
	}
}

func TestDialerErr(t *testing.T) {
	d := netlisten.NewDialer(
		net.Dialer{},
		limiter.NewController(Interval, Ticks, 0, 0),
		limiter.NewController(Interval, Ticks, 0, 0),
	)
	_, err := d.Dial("noone", "noone")
	if err == nil {
		t.Errorf("expected error, got %v", err)
	}
}

func TestListenerErr(t *testing.T) {
	l := netlisten.NewListener(
		&errListener{},
		limiter.NewController(Interval, Ticks, 0, 0),
		limiter.NewController(Interval, Ticks, 0, 0),
	)

	_, err := l.Accept()
	if err == nil || err != errListenerTest {
		t.Errorf("expected %v, got %v", errListenerTest, err)
	}
}

func TestDeadline(t *testing.T) {
	l := listen()

	go func() {
		server := accept(l, new(int64))
		go server.read()
		_, err := io.Copy(server.conn, &noOpReader{})
		panic(err)
	}()

	d := netlisten.NewDialer(
		net.Dialer{},
		limiter.NewController(Interval, Ticks, 1, 1),
		limiter.NewController(Interval, Ticks, 1, 1),
	)
	conn, err := d.Dial(l.Addr().Network(), l.Addr().String())
	if err != nil {
		panic(err)
	}

	timeout := time.Second * 2
	startTime := time.Now()
	conn.(*netlisten.Conn).SetDeadline(startTime.Add(timeout))

	_, err = conn.Write(make([]byte, 1000))
	if err == nil || !xerrors.Is(err, readwrite.ErrDeadline) {
		t.Errorf("expected %v, got %v", readwrite.ErrDeadline, err)
	}

	spent := time.Now().Sub(startTime)
	if r := float64(spent-timeout) / float64(timeout); math.Abs(r) > 0.01 {
		t.Errorf("expected %v, got %v (%f)", timeout, spent, r)
	}

	startTime = time.Now()
	conn.(*netlisten.Conn).SetDeadline(startTime.Add(timeout))

	conn.Read(make([]byte, 1000))
	_, err = conn.Read(make([]byte, 1000))
	if err == nil || !xerrors.Is(err, readwrite.ErrDeadline) {
		t.Errorf("expected %v, got %v", readwrite.ErrDeadline, err)
	}

	spent = time.Now().Sub(startTime)
	if r := float64(spent-timeout) / float64(timeout); math.Abs(r) > 0.01 {
		t.Errorf("expected %v, got %v (%f)", timeout, spent, r)
	}

	startTime = time.Now()
	conn.(*netlisten.Conn).SetWriteDeadline(startTime.Add(timeout))

	_, err = conn.Write(make([]byte, 1000))
	if err == nil || !xerrors.Is(err, readwrite.ErrDeadline) {
		t.Errorf("expected %v, got %v", readwrite.ErrDeadline, err)
	}

	spent = time.Now().Sub(startTime)
	if r := float64(spent-timeout) / float64(timeout); math.Abs(r) > 0.01 {
		t.Errorf("expected %v, got %v (%f)", timeout, spent, r)
	}

	startTime = time.Now()
	conn.(*netlisten.Conn).SetReadDeadline(startTime.Add(timeout))

	_, err = conn.Read(make([]byte, 1000))
	if err == nil || !xerrors.Is(err, readwrite.ErrDeadline) {
		t.Errorf("expected %v, got %v", readwrite.ErrDeadline, err)
	}

	spent = time.Now().Sub(startTime)
	if r := float64(spent-timeout) / float64(timeout); math.Abs(r) > 0.01 {
		t.Errorf("expected %v, got %v (%f)", timeout, spent, r)
	}
}

type noOpReader struct{}

func (r *noOpReader) Read(p []byte) (int, error) {
	return len(p), nil
}

var errListenerTest = xerrors.New("test")

type errListener struct {
	net.Listener
}

func (l *errListener) Accept() (net.Conn, error) {
	return nil, errListenerTest
}
