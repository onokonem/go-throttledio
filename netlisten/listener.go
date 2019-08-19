package netlisten

import (
	"net"
	"time"

	"github.com/onokonem/go-throttledio/limiter"
	"github.com/onokonem/go-throttledio/readwrite"
)

// LimitListener creates a Listener with the per-server and per-connection read limits provided.
// write will be unlimited.
// interval will be 10 seconds, divided to 100 ticks.
func LimitListener(l net.Listener, globalLimit int, connectionLimit int) net.Listener {
	return NewListener(
		l,
		limiter.NewController(10*time.Second, 100, int64(globalLimit), int64(connectionLimit)),
		limiter.NewController(10*time.Second, 100, 0, 0),
	)
}

// Listener is a net.Listener wrapper
type Listener struct {
	net.Listener
	readLimiter  *limiter.Controller
	writeLimiter *limiter.Controller
}

// NewListener creates a Listener
func NewListener(
	listener net.Listener,
	readLimiter *limiter.Controller,
	writeLimiter *limiter.Controller,
) *Listener {
	return &Listener{
		Listener:     listener,
		readLimiter:  readLimiter,
		writeLimiter: writeLimiter,
	}
}

// Accept waits for and returns the next connection to the listener.
func (l *Listener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	var (
		readLimiter  = l.readLimiter.BornLimiter()
		writeLimiter = l.writeLimiter.BornLimiter()
	)

	return &Conn{
			Conn:         conn,
			readLimiter:  readLimiter,
			writeLimiter: writeLimiter,
			r:            readwrite.NewReader(conn, readLimiter, false),
			w:            readwrite.NewWriter(conn, writeLimiter, false),
		},
		nil
}

// ReadLimiter returns a limiter for read.
func (l *Listener) ReadLimiter() *limiter.Controller {
	return l.readLimiter
}

// WriteLimiter returns a limiter for write.
func (l *Listener) WriteLimiter() *limiter.Controller {
	return l.writeLimiter
}
