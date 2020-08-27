// Package blinkm implements a driver for the BlinkM I2C RGB LED.
//
// Datasheet: http://thingm.com/fileadmin/thingm/downloads/BlinkM_datasheet.pdf
//
package blinkm // import "tinygo.org/x/drivers/blinkm"

import "tinygo.org/x/drivers"

// Device wraps an I2C connection to a BlinkM device.
type Device struct {
	bus     drivers.I2C
	Address uint16
}

// New creates a new BlinkM connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{bus, Address}
}

// Configure sets up the device for communication
func (d *Device) Configure() {
	d.bus.Tx(d.Address, []byte{'o'}, nil)
}

// Version returns the version of firmware on the BlinkM.
func (d Device) Version() (major, minor byte, err error) {
	version := []byte{0, 0}
	d.bus.Tx(d.Address, []byte{GET_FIRMWARE}, version)
	return version[0], version[1], nil
}

// SetRGB sets the RGB color on the BlinkM.
func (d Device) SetRGB(r, g, b byte) error {
	d.bus.Tx(d.Address, []byte{TO_RGB, r, g, b}, nil)
	return nil
}

// GetRGB gets the current RGB color on the BlinkM.
func (d Device) GetRGB() (r, g, b byte, err error) {
	color := []byte{0, 0, 0}
	d.bus.Tx(d.Address, []byte{GET_RGB}, color)
	return color[0], color[1], color[2], nil
}

// FadeToRGB sets the RGB color on the BlinkM by fading from the current color
// to the new color.
func (d Device) FadeToRGB(r, g, b byte) error {
	d.bus.Tx(d.Address, []byte{FADE_TO_RGB, r, g, b}, nil)
	return nil
}

// StopScript stops whatever script is currently running on the BlinkM.
func (d Device) StopScript() error {
	d.bus.Tx(d.Address, []byte{STOP_SCRIPT}, nil)
	return nil
}
