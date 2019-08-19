package netlisten

import (
	"context"
	"net"

	"github.com/onokonem/go-throttledio/limiter"
	"github.com/onokonem/go-throttledio/readwrite"
)

// Dialer is a net.Dialer wrapper
type Dialer struct {
	net.Dialer
	readLimiter  *limiter.Controller
	writeLimiter *limiter.Controller
}

// NewDialer creates a Dialer
func NewDialer(
	dialer net.Dialer,
	readLimiter *limiter.Controller,
	writeLimiter *limiter.Controller,
) *Dialer {
	return &Dialer{
		Dialer:       dialer,
		readLimiter:  readLimiter,
		writeLimiter: writeLimiter,
	}
}

// Dial is a wrapper around DialContext().
// Returns a trottler-powered net.Conn
func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

// DialContext is a wrapper around d.Dialer.DialContext().
// Returns a trottler-powered net.Conn
func (d *Dialer) DialContext(ctx context.Context, network, address string) (*Conn, error) {
	conn, err := d.Dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}

	var (
		readLimiter  = d.readLimiter.BornLimiter()
		writeLimiter = d.writeLimiter.BornLimiter()
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
