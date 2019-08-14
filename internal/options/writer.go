// Package options comment should be of this form
package options

// Discard exported type should have comment or be unexported
type Discard struct {
	V bool
}

// ItIsAWriterOption exported func should have comment or be unexported
func (o *Discard) ItIsAWriterOption() {}
