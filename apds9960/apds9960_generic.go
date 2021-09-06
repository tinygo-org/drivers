// +build !nano_33_ble

package apds9960

import "tinygo.org/x/drivers"

// New creates a new APDS-9960 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C, deviceType uint8) Device {
	return Device{bus: bus, Address: ADPS9960_ADDRESS, mode: MODE_NONE}
}
