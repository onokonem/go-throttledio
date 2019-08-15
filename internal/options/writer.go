// Package options comment should be of this form
package options

// Discard exported type should have comment or be unexported
type Discard struct {
	Discard bool
	NoErr   bool
}

// ItIsAWriterOption exported func should have comment or be unexported
func (o *Discard) ItIsAWriterOption() {}
