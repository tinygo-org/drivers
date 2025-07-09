//go:build baremetal

package epd2in66b

import "machine"

type Config struct {
	ResetPin      machine.Pin
	DataPin       machine.Pin
	ChipSelectPin machine.Pin
	BusyPin       machine.Pin
}

// Configure configures the device and its pins.
func (d *Device) Configure(c Config) error {
	cs := c.ChipSelectPin
	dc := c.DataPin
	rst := c.ResetPin
	busy := c.BusyPin

	cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	dc.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rst.Configure(machine.PinConfig{Mode: machine.PinOutput})
	busy.Configure(machine.PinConfig{Mode: machine.PinInput})
	d.cs = cs.Set
	d.dc = dc.Set
	d.rst = rst.Set
	d.busy = busy.Get
	return nil
}
