package drivers

// Pin types abstract the underlying implementation of a pin,
// allowing the use of different hardware or libraries without changing driver code.
//
// Note, pin mode functionality is not part of the Pin interface.
// Implementations must ensure correct pin mode is set when Get or Set methods are called.
//
// Drivers must use SafePin(), SafePinInput() and SafePinOutput() wrappers in constructors.
// The client code can pass either machine.Pin or a custom implementation of the Pin interface.
// Wrappers for TinyGo's machine.Pin are provided in pin_tinygo.go.
// It's custom implementation's responsibility to configure the pin modes correctly as
// Wrappers in pin_generic.go are no-ops.

// PinInput is an interface for reading the state of a pin.
type PinInput interface {
	Get() bool
}

// PinOutput is an interface for setting the state of a pin.
type PinOutput interface {
	Set(bool)
}

// Pin is an interface for a pin that can be both input and output.
type Pin interface {
	PinInput
	PinOutput
}

// PinGet is a function type for getting the state of a pin
// Shall be used in high-performance drivers to avoid interface call overhead.
type PinGet func() bool

// PinSet is a function type for setting the state of a pin.
// Shall be used in high-performance drivers to avoid interface call overhead.
type PinSet func(bool)
