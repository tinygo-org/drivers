//go:build nano_33_ble

// Nano 33 BLE [Sense] has LSM9DS1 unit on-board.
// This custom Configure function powers unit up
// and enables I2C, so unit can can be accessed.
package lsm9ds1

import (
	"machine"
	"time"
)

// Configure sets up the device for communication.
func (d *Device) Configure(cfg Configuration) error {
	// Following lines are Nano 33 BLE specific, they have nothing to do with sensor per se
	machine.LSM_PWR.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.LSM_PWR.High()
	machine.I2C_PULLUP.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.I2C_PULLUP.High()
	// Wait a moment
	time.Sleep(100 * time.Millisecond)
	// Common initialisation code
	return d.doConfigure(cfg)
}
