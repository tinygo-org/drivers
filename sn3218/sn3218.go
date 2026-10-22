// Package sn3218 implements a driver for the SI-EN SN3218 18 channel LED
// driver, as used to drive the RGB backlight on the Pimoroni Display-o-Tron
// HAT.
//
// Datasheet: https://www.si-en.com/uploadpdf/s2011528172924.pdf
//
// This driver ports the register sequence used by the Pimoroni Python
// driver: https://github.com/pimoroni/sn3218
package sn3218 // import "tinygo.org/x/drivers/sn3218"

import (
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// DefaultAddress is the fixed I2C address of the SN3218.
const DefaultAddress = 0x54

// NumChannels is the number of PWM channels the SN3218 drives.
const NumChannels = 18

const (
	regEnableOutput = 0x00
	regSetPWMValues = 0x01
	regEnableLEDs   = 0x13
	regUpdate       = 0x16
	regReset        = 0x17
)

// Device is a handle to a SN3218 LED driver.
type Device struct {
	bus     drivers.I2C
	Address uint8
}

// New returns a new sn3218 driver on the given I2C bus, using the fixed
// device address.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: DefaultAddress,
	}
}

// Configure resets the chip, enables all 18 channels and turns the output
// stage on.
func (d *Device) Configure() error {
	if err := d.Reset(); err != nil {
		return err
	}
	if err := d.EnableChannels(0x3FFFF); err != nil {
		return err
	}
	return d.Enable()
}

// Enable turns the output stage on.
func (d *Device) Enable() error {
	return legacy.WriteRegister(d.bus, d.Address, regEnableOutput, []byte{0x01})
}

// Disable turns the output stage off.
func (d *Device) Disable() error {
	return legacy.WriteRegister(d.bus, d.Address, regEnableOutput, []byte{0x00})
}

// Reset restores all internal registers to their power on state.
func (d *Device) Reset() error {
	return legacy.WriteRegister(d.bus, d.Address, regReset, []byte{0xFF})
}

// EnableChannels enables or disables each of the 18 channels. Bit n of mask
// controls channel n, 1 turns the channel on, 0 turns it off. Channels must
// be enabled here before SetChannels has any effect on them.
func (d *Device) EnableChannels(mask uint32) error {
	data := []byte{
		uint8(mask & 0x3F),
		uint8((mask >> 6) & 0x3F),
		uint8((mask >> 12) & 0x3F),
	}
	if err := legacy.WriteRegister(d.bus, d.Address, regEnableLEDs, data); err != nil {
		return err
	}
	return d.latch()
}

// SetChannels writes a raw PWM value, 0 to 255, to each of the 18 channels
// and latches the change so it takes effect.
func (d *Device) SetChannels(values [NumChannels]uint8) error {
	if err := legacy.WriteRegister(d.bus, d.Address, regSetPWMValues, values[:]); err != nil {
		return err
	}
	return d.latch()
}

// latch tells the chip to move the values written by SetChannels or
// EnableChannels from its shadow registers into the active PWM registers.
func (d *Device) latch() error {
	return legacy.WriteRegister(d.bus, d.Address, regUpdate, []byte{0xFF})
}
