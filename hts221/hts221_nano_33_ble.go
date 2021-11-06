// +build nano_33_ble

package hts221

import (
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

// New creates a new HTS221 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	// turn on internal power pin (machine.P0_22) and I2C1 pullups power pin (machine.P1_00)
	// and wait a moment.
	ENV := machine.P0_22
	ENV.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ENV.High()
	R := machine.P1_00
	R.Configure(machine.PinConfig{Mode: machine.PinOutput})
	R.High()
	time.Sleep(time.Millisecond * 10)

	return Device{bus: bus, Address: HTS221_ADDRESS}
}
