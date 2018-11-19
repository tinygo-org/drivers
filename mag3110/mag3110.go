// Package mag3110 implements a driver for the MAG3110 3-axis magnetometer by
// Freescale/NXP.
//
// Datasheet: https://www.nxp.com/docs/en/data-sheet/MAG3110.pdf
package mag3110

import (
	"machine"
)

// Device wraps an I2C connection to a MAG3110 device.
type Device struct {
	bus machine.I2C
}

// New creates a new MAG3110 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus machine.I2C) Device {
	return Device{bus}
}

// Connected returns whether a MAG3110 has been found.
// It does a "who am I" request and checks the response.
func (d Device) Connected() bool {
	data := []byte{0}
	d.bus.ReadRegister(Address, WHO_AM_I, data)
	return data[0] == 0xC4
}

// Configure sets up the device for communication.
func (d Device) Configure() {
	d.bus.WriteRegister(Address, CTRL_REG2, []uint8{0x80}) // Power down when not used
}

// ReadMagnetic reads the vectors of the magnetic field of the device and
// returns it.
func (d Device) ReadMagnetic() (x int16, y int16, z int16) {
	d.bus.WriteRegister(Address, CTRL_REG1, []uint8{0x1a}) // Request a measurement

	data := make([]byte, 6)
	d.bus.ReadRegister(Address, OUT_X_MSB, data)
	x = int16((uint16(data[0]) << 8) | uint16(data[1]))
	y = int16((uint16(data[2]) << 8) | uint16(data[3]))
	z = int16((uint16(data[4]) << 8) | uint16(data[5]))
	return
}

// ReadTemperature reads the current die temperature in degrees Celsius.
func (d Device) ReadTemperature() (temp int8) {
	data := make([]byte, 1)
	d.bus.ReadRegister(Address, DIE_TEMP, data)
	return int8(data[0])
}
