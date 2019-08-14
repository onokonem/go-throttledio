package throttledio

import (
	"time"

	"github.com/onokonem/go-throttledio/internal/options"
)

type WriterOption interface {
	ItIsAWriterOption()
}

type ReaderOption interface {
	ItIsAReaderOption()
}

type CommonOption interface {
	ItIsAWriterOption()
	ItIsAReaderOption()
}

func SetInterval(v time.Duration) CommonOption {
	return &options.Interval{V: v}
}

func SetTicks(v uint) CommonOption {
	return &options.Ticks{V: v}
}

func SetNoError(v bool) WriterOption {
	return &options.NoError{V: v}
}

func SetSpeed(v float64) CommonOption {
	return &options.Speed{V: v}
}

func SetDiscard(v bool) WriterOption {
	return &options.Discard{V: v}
}
