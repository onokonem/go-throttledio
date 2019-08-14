package options

import "time"

type Interval struct {
	V time.Duration
}

func (o *Interval) ItIsAWriterOption() {}
func (o *Interval) ItIsAReaderOption() {}

type Ticks struct {
	V uint
}

func (o *Ticks) ItIsAWriterOption() {}
func (o *Ticks) ItIsAReaderOption() {}

type Speed struct {
	V float64
}

func (o *Speed) ItIsAWriterOption() {}
func (o *Speed) ItIsAReaderOption() {}

type NoError struct {
	V bool
}

func (o *NoError) ItIsAWriterOption() {}
func (o *NoError) ItIsAReaderOption() {}
