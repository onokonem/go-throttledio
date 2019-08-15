// Package options comment should be of this form
package options

import "time"

// Interval exported type should have comment or be unexported
type Interval struct {
	Interval time.Duration
	Ticks    uint
}

// ItIsAWriterOption exported func should have comment or be unexported
func (o *Interval) ItIsAWriterOption() {}

// ItIsAReaderOption exported func should have comment or be unexported
func (o *Interval) ItIsAReaderOption() {}

// Speed exported type should have comment or be unexported
type Speed struct {
	V float64
}

// ItIsAWriterOption exported func should have comment or be unexported
func (o *Speed) ItIsAWriterOption() {}

// ItIsAReaderOption exported func should have comment or be unexported
func (o *Speed) ItIsAReaderOption() {}
