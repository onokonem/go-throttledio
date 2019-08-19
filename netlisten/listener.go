package netlisten

import (
	"net"

	"github.com/onokonem/go-throttledio/limiter"
	"github.com/onokonem/go-throttledio/readwrite"
)

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
