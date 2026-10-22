//go:build tinygo

package st7036

import (
	"machine"

	"tinygo.org/x/drivers"
)

// New returns a driver with an SPI bus that is already configured.
// rs selects commands or data. Use machine.NoPin if reset is not connected.
func New(bus drivers.SPI, cs, rs, reset machine.Pin) Device {
	cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	cs.High()
	rs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d := Device{
		bus:                    bus,
		cs:                     cs,
		rs:                     rs,
		instructionSetTemplate: defaultInstructionSetTemplate,
	}
	if reset != machine.NoPin {
		reset.Configure(machine.PinConfig{Mode: machine.PinOutput})
		reset.High()
		d.reset = reset
	}
	return d
}
