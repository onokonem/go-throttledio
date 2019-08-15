package throttledio

import (
	"time"

	"github.com/onokonem/go-throttledio/internal/options"
)

// WriterOption is an interface option has to implement to passed to NewWriter.
type WriterOption interface {
	ItIsAWriterOption()
}

// ReaderOption is an interface option has to implement to passed to NewReader.
type ReaderOption interface {
	ItIsAReaderOption()
}

// CommonOption could be passed to NewReader and to NewWriter.
type CommonOption interface {
	ItIsAWriterOption()
	ItIsAReaderOption()
}

// SetInterval creates an option to set measuring interval and a number of gaps this interval will be divided to.
func SetInterval(interval time.Duration, ticks uint) CommonOption {
	return &options.Interval{Interval: interval, Ticks: ticks}
}

// SetNoError creates an option controlling ErrExceeded error.
// If true ErrExceeded will be returned as soon as Writer
// would not be able to complete a write request because of throttling.
func SetNoError(v bool) WriterOption {
	return &options.NoError{V: v}
}

// SetSpeed c=reates an option to set the CPS.
func SetSpeed(v float64) CommonOption {
	return &options.Speed{V: v}
}

// SetDiscard creates an option to control the throttling policy. If true the throttled data will be discarded.
func SetDiscard(v bool) WriterOption {
	return &options.Discard{V: v}
}
