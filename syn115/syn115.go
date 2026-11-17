// Package syn115 provides a driver for the SYN113/SYN115 ASK transmitter.
//
// Datasheet: https://www.openimpulse.com/blog/wp-content/uploads/wpsc/downloadables/SYN115-ASK-Transmitter-Datasheet.pdf
package syn115

import "machine"

// Device is a SYN115 transmitter.
type Device struct {
	pin machine.Pin
}

// New returns a new SYN115 device connected to the ASK data pin.
func New(pin machine.Pin) Device {
	return Device{pin: pin}
}

// Configure sets the ASK data pin as an output and turns the carrier off.
func (d Device) Configure() {
	d.pin.Configure(
		machine.PinConfig{
			Mode: machine.PinOutput,
		},
	)
	d.pin.Set(false)
}

// Set turns the carrier on when high and off when low.
func (d Device) Set(high bool) {
	d.pin.Set(high)
}
