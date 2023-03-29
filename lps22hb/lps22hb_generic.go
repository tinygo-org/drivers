//go:build !nano_33_ble

package lps22hb

import "tinygo.org/x/drivers"

// Configure sets up the LPS22HB device for communication.
func (d *Device) Configure() {
	// set to block update mode
	d.bus.WriteRegister(d.Address, LPS22HB_CTRL1_REG, []byte{0x02})
}
