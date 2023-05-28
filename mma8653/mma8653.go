// Package mma8653 provides a driver for the MMA8653 3-axis accelerometer by
// Freescale/NXP.
//
// Datasheet:
// https://www.nxp.com/docs/en/data-sheet/MMA8653FC.pdf
package mma8653 // import "tinygo.org/x/drivers/mma8653"

import (
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// Device wraps an I2C connection to a MMA8653 device.
type Device struct {
	bus         drivers.I2C
	Address     uint16
	sensitivity Sensitivity
}

// New creates a new MMA8653 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{bus, Address, Sensitivity2G}
}

// Connected returns whether a MMA8653 has been found.
// It does a "who am I" request and checks the response.
func (d Device) Connected() bool {
	data := []byte{0}
	legacy.ReadRegister(d.bus, uint8(d.Address), WHO_AM_I, data)
	return data[0] == 0x5A
}

// Configure sets up the device for communication.
func (d *Device) Configure(speed DataRate, sensitivity Sensitivity) error {
	// Set mode to STANDBY to be able to change the sensitivity.
	err := legacy.WriteRegister(d.bus, uint8(d.Address), CTRL_REG1, []uint8{0})
	if err != nil {
		return err
	}

	// Set sensitivity (2G, 4G, 8G).
	err = legacy.WriteRegister(d.bus, uint8(d.Address), XYZ_DATA_CFG, []uint8{uint8(sensitivity)})
	if err != nil {
		return err
	}
	d.sensitivity = sensitivity

	// Set mode to ACTIVE and set the data rate.
	err = legacy.WriteRegister(d.bus, uint8(d.Address), CTRL_REG1, []uint8{(uint8(speed) << 3) | 1})
	if err != nil {
		return err
	}
	return nil
}

// ReadAcceleration reads the current acceleration from the device and returns
// it in µg (micro-gravity). When one of the axes is pointing straight to Earth
// and the sensor is not moving the returned value will be around 1000000 or
// -1000000.
func (d Device) ReadAcceleration() (x int32, y int32, z int32, err error) {
	data := make([]byte, 6)
	err = legacy.ReadRegister(d.bus, uint8(d.Address), OUT_X_MSB, data)
	shift := uint32(8)
	switch d.sensitivity {
	case Sensitivity4G:
		shift = 7
	case Sensitivity8G:
		shift = 6
	}
	x = int32(int16((uint16(data[0])<<8)|uint16(data[1]))) * 15625 >> shift
	y = int32(int16((uint16(data[2])<<8)|uint16(data[3]))) * 15625 >> shift
	z = int32(int16((uint16(data[4])<<8)|uint16(data[5]))) * 15625 >> shift
	return
}
