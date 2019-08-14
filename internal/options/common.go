// Package options comment should be of this form
package options

import "time"

// Interval exported type should have comment or be unexported
type Interval struct {
	V time.Duration
}

// ItIsAWriterOption exported func should have comment or be unexported
func (o *Interval) ItIsAWriterOption() {}

// ItIsAReaderOption exported func should have comment or be unexported
func (o *Interval) ItIsAReaderOption() {}

// Ticks exported type should have comment or be unexported
type Ticks struct {
	V uint
}

// ItIsAWriterOption exported func should have comment or be unexported
func (o *Ticks) ItIsAWriterOption() {}

// ItIsAReaderOption exported func should have comment or be unexported
func (o *Ticks) ItIsAReaderOption() {}

// Speed exported type should have comment or be unexported
type Speed struct {
	V float64
}

// ItIsAWriterOption exported func should have comment or be unexported
func (o *Speed) ItIsAWriterOption() {}

// ItIsAReaderOption exported func should have comment or be unexported
func (o *Speed) ItIsAReaderOption() {}

// NoError exported type should have comment or be unexported
type NoError struct {
	V bool
}

// ItIsAWriterOption exported func should have comment or be unexported
func (o *NoError) ItIsAWriterOption() {}

// ItIsAReaderOption exported func should have comment or be unexported
func (o *NoError) ItIsAReaderOption() {}
