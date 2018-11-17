// Package mpu6050 provides a driver for the MPU6050 accelerometer and gyroscope
// made by InvenSense.
//
// Datasheets:
// https://store.invensense.com/datasheets/invensense/MPU-6050_DataSheet_V3%204.pdf
// https://www.invensense.com/wp-content/uploads/2015/02/MPU-6000-Register-Map1.pdf
package mpu6050

import (
	"machine"
)

// Device wraps an I2C connection to a MPU6050 device.
type Device struct {
	bus machine.I2C
}

// NewI2C creates a new MPU6050 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func NewI2C(bus machine.I2C) Device {
	return Device{bus}
}

// Connected returns whether a MPU6050 has been found.
// It does a "who am I" request and checks the response.
func (d Device) Connected() bool {
	data := []byte{0}
	d.bus.ReadRegister(Address, WHO_AM_I, data)
	return data[0] == 0x68
}

// Configure sets up the device for communication.
func (d Device) Configure() {
	d.bus.WriteRegister(Address, PWR_MGMT_1, []uint8{0})
}

// ReadAcceleration reads the current acceleration from the device and returns
// it.
func (d Device) ReadAcceleration() (x int16, y int16, z int16) {
	data := make([]byte, 6)
	d.bus.ReadRegister(Address, ACCEL_XOUT_H, data)
	x = int16((uint16(data[0]) << 8) | uint16(data[1]))
	y = int16((uint16(data[2]) << 8) | uint16(data[3]))
	z = int16((uint16(data[4]) << 8) | uint16(data[5]))
	return
}

// ReadRotation reads the current rotation from the device and returns it.
func (d Device) ReadRotation() (x int16, y int16, z int16) {
	data := make([]byte, 6)
	d.bus.ReadRegister(Address, GYRO_XOUT_H, data)
	x = int16((uint16(data[0]) << 8) | uint16(data[1]))
	y = int16((uint16(data[2]) << 8) | uint16(data[3]))
	z = int16((uint16(data[4]) << 8) | uint16(data[5]))
	return
}
