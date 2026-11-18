// Package syn480r provides a driver for the SYN470R/SYN480R ASK receiver.
//
// Datasheet: https://www.lcsc.com/datasheet/C916347.pdf
package syn480r

import "machine"

// Device is a SYN480R receiver.
type Device struct {
	pin    machine.Pin
	invert bool
}

// Config holds the receiver options.
type Config struct {
	// Invert flips the data output polarity. Use it for modules that output
	// low while the carrier is present.
	Invert bool
}

// New returns a new SYN480R device connected to the data output pin.
func New(pin machine.Pin) Device {
	return NewWithConfig(pin, Config{})
}

// NewWithConfig is New with the options in config.
func NewWithConfig(pin machine.Pin, config Config) Device {
	return Device{
		pin:    pin,
		invert: config.Invert,
	}
}

// Configure sets the data output pin as an input.
func (d Device) Configure() {
	d.pin.Configure(machine.PinConfig{Mode: machine.PinInput})
}

// Get returns the demodulated level, high when a carrier is present.
func (d Device) Get() bool {
	return d.pin.Get() != d.invert
}
