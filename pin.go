package drivers

// Pin types abstract the underlying implementation of a pin,
// allowing the use of different hardware or libraries without changing driver code.
//
// Note, pin mode functionality is not part of the Pin interface.
// Client code is responsible for configuring pin modes correctly.
// This can be done either before passing the pin to a driver constructor
// or by ensuring correct pin mode is set when Get or Set methods are called.
// See rpio package for an example of a pin implementation that does this.
//
// Drivers that used to configure output pin mode in their constructors can use
// legacy.PinOutput() wrapper to keep the same behavior for machine.Pin.
// See internal/legacy/pinlegacy_tinygo.go and internal/legacy/pinlegacy_generic.go for details.
//
// All new drivers are encouraged to not configure pin modes in their constructors and
// do not depend on either machine package or legacy wrappers.

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
