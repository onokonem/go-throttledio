package options

type Discard struct {
	V bool
}

func (o *Discard) ItIsAWriterOption() {}
