package netlisten

import (
	"net"
	"time"

	"github.com/onokonem/go-throttledio/limiter"
	"github.com/onokonem/go-throttledio/readwrite"
)

// Conn is a net.Conn implementation powered with throttling.
type Conn struct {
	net.Conn
	readLimiter  *limiter.Limiter
	writeLimiter *limiter.Limiter
	w            *readwrite.Writer
	r            *readwrite.Reader
}

// SetReadCPS sets the read limit.
func (c *Conn) SetReadCPS(cps int64) {
	c.readLimiter.SetCPS(cps)
}

// SetWriteCPS sets the write limit.
func (c *Conn) SetWriteCPS(cps int64) {
	c.writeLimiter.SetCPS(cps)
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *Conn) Read(b []byte) (n int, err error) {
	return c.r.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *Conn) Write(b []byte) (n int, err error) {
	return c.w.Write(b)
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (c *Conn) SetDeadline(t time.Time) error {
	c.r.SetDeadline(t)
	c.w.SetDeadline(t)
	return c.Conn.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c *Conn) SetReadDeadline(t time.Time) error {
	c.r.SetDeadline(t)
	return c.Conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	c.w.SetDeadline(t)
	return c.Conn.SetWriteDeadline(t)
}
