//go:build nano_33_ble

package lps22hb

import (
	"machine"
	"time"
)

// Configure sets up the LPS22HB device for communication.
func (d *Device) Configure() {
	// Following lines are Nano 33 BLE specific, they have nothing to do with sensor per se
	machine.LPS_PWR.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.LPS_PWR.High()
	machine.I2C_PULLUP.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.I2C_PULLUP.High()
	// Wait a moment
	time.Sleep(10 * time.Millisecond)

	// set to block update mode
	d.bus.WriteRegister(d.Address, LPS22HB_CTRL1_REG, []byte{0x02})
}
