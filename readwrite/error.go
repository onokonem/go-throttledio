package readwrite

import (
	"errors"
	"net"
)

// Errors
var (
	ErrExceeded = &Error{errors.New("bandwidth exceeded"), false, true}
	ErrDeadline = &Error{errors.New("deadline reached"), true, false}
)

var (
	_ error     = (*Error)(nil)
	_ net.Error = (*Error)(nil)
)

// An Error represents a readwrite error.
type Error struct {
	error
	timeout   bool
	temporary bool
}

// Timeout flag
func (e *Error) Timeout() bool { return e.timeout }

// Temporary flag
func (e *Error) Temporary() bool { return e.temporary }
