package readwrite

import "golang.org/x/xerrors"

// Errors
var (
	ErrExceeded = xerrors.New("bandwidth exceeded")
	ErrDeadline = xerrors.New("deadline reached")
)
