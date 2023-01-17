//go:build nano_33_ble

package apds9960

import (
	"machine"
	"time"
)

// Configure sets up the APDS-9960 device.
func (d *Device) Configure(cfg Configuration) {

	// Following lines are Nano 33 BLE specific, they have nothing to do with sensor per se
	machine.LSM_PWR.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.LSM_PWR.High()
	machine.I2C_PULLUP.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.I2C_PULLUP.High()
	// Wait a moment
	time.Sleep(10 * time.Millisecond)

	// configure device
	d.configureDevice(cfg)
}
