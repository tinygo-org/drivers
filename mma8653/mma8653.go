// Package mma8653 provides a driver for the MMA8653 3-axis accelerometer by
// Freescale/NXP.
//
// Datasheet:
// https://www.nxp.com/docs/en/data-sheet/MMA8653FC.pdf
package mma8653

import (
	"machine"
)

// Device wraps an I2C connection to a MMA8653 device.
type Device struct {
	bus machine.I2C
}

// NewI2C creates a new MMA8653 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func NewI2C(bus machine.I2C) Device {
	return Device{bus}
}

// Connected returns whether a MMA8653 has been found.
// It does a "who am I" request and checks the response.
func (d Device) Connected() bool {
	data := []byte{0}
	d.bus.ReadRegister(Address, WHO_AM_I, data)
	return data[0] == 0x5A
}

// Configure sets up the device for communication.
func (d Device) Configure(speed DataRate) {
	data := (uint8(speed) << 3) | 1 // set data rate and ACTIVE mode
	d.bus.WriteRegister(Address, CTRL_REG1, []uint8{data})
}

// ReadOrientation reads the current orientation from the device and returns it.
func (d Device) ReadOrientation() (x int16, y int16, z int16) {
	data := make([]byte, 6)
	d.bus.ReadRegister(Address, OUT_X_MSB, data)
	x = int16((uint16(data[0]) << 8) | uint16(data[1]))
	y = int16((uint16(data[2]) << 8) | uint16(data[3]))
	z = int16((uint16(data[4]) << 8) | uint16(data[5]))
	return
}
