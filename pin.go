package drivers

// Pin is a digital pin interface.
// It is notably implemented by the machine.Pin type.
type Pin interface {
	PinIn
	PinOut
}

type PinIn interface {
	Get() bool
}

type PinOut interface {
	High() // deprecated: use Set(true)
	Low()  // deprecated: use Set(false)
	Set(high bool)
}
