//go:build !xiao_ble

package lsm6ds3tr

// Configure sets up the device for communication.
func (d *Device) Configure(cfg Configuration) error {
	return d.doConfigure(cfg)
}
