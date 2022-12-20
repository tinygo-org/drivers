//go:build !nano_33_ble

package lsm9ds1

// Configure sets up the device for communication.
func (d *Device) Configure(cfg Configuration) error {
	return d.doConfigure(cfg)
}
