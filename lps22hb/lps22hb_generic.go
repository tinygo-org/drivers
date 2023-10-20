//go:build !nano_33_ble

package lps22hb

import "tinygo.org/x/drivers/internal/legacy"

// Configure sets up the LPS22HB device for communication.
func (d *Device) Configure() {
	// set to block update mode
	legacy.WriteRegister(d.bus, d.Address, LPS22HB_CTRL1_REG, []byte{0x02})
}
