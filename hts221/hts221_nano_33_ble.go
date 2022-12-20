//go:build nano_33_ble

package hts221

import (
	"machine"
	"time"
)

// Configure sets up the HTS221 device for communication.
func (d *Device) Configure() {
	// Following lines are Nano 33 BLE specific, they have nothing to do with sensor per se
	machine.HTS_PWR.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.HTS_PWR.High()
	machine.I2C_PULLUP.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.I2C_PULLUP.High()
	// Wait a moment
	time.Sleep(10 * time.Millisecond)

	// read calibration data
	d.calibration()
	// activate device and use block data update mode
	d.Power(true)
}
