//go:build !nano_33_ble
// +build !nano_33_ble

package apds9960

import "tinygo.org/x/drivers"

// Configure sets up the APDS-9960 device.
func (d *Device) Configure(cfg Configuration) {
	// configure device
	d.configureDevice(cfg)
}
