// Package mag3110 implements a driver for the MAG3110 3-axis magnetometer by
// Freescale/NXP.
//
// Datasheet: https://www.nxp.com/docs/en/data-sheet/MAG3110.pdf
package mag3110 // import "tinygo.org/x/drivers/mag3110"

import (
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// Device wraps an I2C connection to a MAG3110 device.
type Device struct {
	bus     drivers.I2C
	Address uint16
}

// New creates a new MAG3110 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{bus, Address}
}

// Connected returns whether a MAG3110 has been found.
// It does a "who am I" request and checks the response.
func (d Device) Connected() bool {
	data := []byte{0}
	legacy.ReadRegister(d.bus, uint8(d.Address), WHO_AM_I, data)
	return data[0] == 0xC4
}

// Configure sets up the device for communication.
func (d Device) Configure() {
	legacy.WriteRegister(d.bus, uint8(d.Address), CTRL_REG2, []uint8{0x80}) // Power down when not used
}

// ReadMagnetic reads the vectors of the magnetic field of the device and
// returns it.
func (d Device) ReadMagnetic() (x int16, y int16, z int16) {
	legacy.WriteRegister(d.bus, uint8(d.Address), CTRL_REG1, []uint8{0x1a}) // Request a measurement

	data := make([]byte, 6)
	legacy.ReadRegister(d.bus, uint8(d.Address), OUT_X_MSB, data)
	x = int16((uint16(data[0]) << 8) | uint16(data[1]))
	y = int16((uint16(data[2]) << 8) | uint16(data[3]))
	z = int16((uint16(data[4]) << 8) | uint16(data[5]))
	return
}

// ReadTemperature reads and returns the current die temperature in
// celsius milli degrees (°C/1000).
func (d Device) ReadTemperature() (int32, error) {
	data := make([]byte, 1)
	legacy.ReadRegister(d.bus, uint8(d.Address), DIE_TEMP, data)
	return int32(data[0]) * 1000, nil
}
