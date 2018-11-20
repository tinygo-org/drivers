// Package blinkm implements a driver for the BlinkM I2C RGB LED.
//
// Datasheet: http://thingm.com/fileadmin/thingm/downloads/BlinkM_datasheet.pdf
package blinkm

import (
	"machine"
)

// Device wraps an I2C connection to a BlinkM device.
type Device struct {
	bus machine.I2C
}

// New creates a new BlinkM connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus machine.I2C) Device {
	return Device{bus}
}

// Version returns the version of firmware on the BlinkM.
func (d Device) Version() ([]byte, error) {
	version := []byte{0, 0}
	d.bus.ReadRegister(Address, GET_FIRMWARE, version)
	return version, nil
}

// SetRGB sets the RGB color on the BlinkM.
func (d Device) SetRGB(r, g, b byte) error {
	d.bus.WriteRegister(Address, TO_RGB, []byte{r, g, b})
	return nil
}

// GetRGB gets the current RGB color on the BlinkM.
func (d Device) GetRGB() (r, g, b byte, err error) {
	color := []byte{0, 0, 0}
	d.bus.ReadRegister(Address, GET_RGB, color)
	return color[0], color[1], color[2], nil
}

// FadeToRGB sets the RGB color on the BlinkM by fading from the current color
// to the new color.
func (d Device) FadeToRGB(r, g, b byte) error {
	d.bus.WriteRegister(Address, FADE_TO_RGB, []byte{r, g, b})
	return nil
}

// StopScript stops whatever script is currently running on the BlinkM.
func (d Device) StopScript() error {
	d.bus.WriteRegister(Address, STOP_SCRIPT, nil)
	return nil
}
