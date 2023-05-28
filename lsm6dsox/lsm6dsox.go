// Package lsm6dsox implements a driver for the LSM6DSOX
// a 6 axis Inertial Measurement Unit (IMU)
//
// Datasheet: https://www.st.com/resource/en/datasheet/lsm6dsox.pdf
package lsm6dsox // import "tinygo.org/x/drivers/lsm6dsox"

import (
	"errors"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

type AccelRange uint8
type AccelSampleRate uint8

type GyroRange uint8
type GyroSampleRate uint8

// Device wraps an I2C connection to a LSM6DSOX device.
type Device struct {
	bus             drivers.I2C
	Address         uint16
	accelMultiplier int32
	gyroMultiplier  int32
	buf             [6]uint8
}

// Configuration for LSM6DSOX device.
type Configuration struct {
	AccelRange      AccelRange
	AccelSampleRate AccelSampleRate
	GyroRange       GyroRange
	GyroSampleRate  GyroSampleRate
}

var errNotConnected = errors.New("lsm6dsox: failed to communicate with acel/gyro sensor")

// New creates a new LSM6DSOX connection. The I2C bus must already be configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) *Device {
	return &Device{
		bus:     bus,
		Address: Address,
	}
}

// Configure sets up the device for communication.
func (d *Device) Configure(cfg Configuration) (err error) {

	// Verify unit communication
	if !d.Connected() {
		return errNotConnected
	}

	// Multipliers come from "Table 2. Mechanical characteristics" of the datasheet * 1000
	switch cfg.AccelRange {
	case ACCEL_2G:
		d.accelMultiplier = 61
	case ACCEL_4G:
		d.accelMultiplier = 122
	case ACCEL_8G:
		d.accelMultiplier = 244
	case ACCEL_16G:
		d.accelMultiplier = 488
	}
	switch cfg.GyroRange {
	case GYRO_250DPS:
		d.gyroMultiplier = 8750
	case GYRO_500DPS:
		d.gyroMultiplier = 17500
	case GYRO_1000DPS:
		d.gyroMultiplier = 35000
	case GYRO_2000DPS:
		d.gyroMultiplier = 70000
	}

	data := d.buf[:1]
	// Configure accelerometer
	data[0] = uint8(cfg.AccelRange) | uint8(cfg.AccelSampleRate)
	err = legacy.WriteRegister(d.bus, uint8(d.Address), CTRL1_XL, data)
	if err != nil {
		return
	}
	// Configure gyroscope
	data[0] = uint8(cfg.GyroRange) | uint8(cfg.GyroSampleRate)
	err = legacy.WriteRegister(d.bus, uint8(d.Address), CTRL2_G, data)
	if err != nil {
		return
	}

	return nil
}

// Connected returns whether a LSM6DSOX has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {
	data := d.buf[:1]
	legacy.ReadRegister(d.bus, uint8(d.Address), WHO_AM_I, data)
	return data[0] == 0x6C
}

// ReadAcceleration reads the current acceleration from the device and returns
// it in µg (micro-gravity). When one of the axes is pointing straight to Earth
// and the sensor is not moving the returned value will be around 1000000 or
// -1000000.
func (d *Device) ReadAcceleration() (x, y, z int32, err error) {
	data := d.buf[:6]
	err = legacy.ReadRegister(d.bus, uint8(d.Address), OUTX_L_A, data)
	if err != nil {
		return
	}
	x = int32(int16((uint16(data[1])<<8)|uint16(data[0]))) * d.accelMultiplier
	y = int32(int16((uint16(data[3])<<8)|uint16(data[2]))) * d.accelMultiplier
	z = int32(int16((uint16(data[5])<<8)|uint16(data[4]))) * d.accelMultiplier
	return
}

// ReadRotation reads the current rotation from the device and returns it in
// µ°/s (micro-degrees/sec). This means that if you were to do a complete
// rotation along one axis and while doing so integrate all values over time,
// you would get a value close to 360000000.
func (d *Device) ReadRotation() (x, y, z int32, err error) {
	data := d.buf[:6]
	err = legacy.ReadRegister(d.bus, uint8(d.Address), OUTX_L_G, data)
	if err != nil {
		return
	}
	x = int32(int16((uint16(data[1])<<8)|uint16(data[0]))) * d.gyroMultiplier
	y = int32(int16((uint16(data[3])<<8)|uint16(data[2]))) * d.gyroMultiplier
	z = int32(int16((uint16(data[5])<<8)|uint16(data[4]))) * d.gyroMultiplier
	return
}

// ReadTemperature returns the temperature in celsius milli degrees (°C/1000)
func (d *Device) ReadTemperature() (t int32, err error) {
	data := d.buf[:2]
	err = legacy.ReadRegister(d.bus, uint8(d.Address), OUT_TEMP_L, data)
	if err != nil {
		return
	}
	// From "Table 4. Temperature sensor characteristics"
	// temp = value/256 + 25
	t = 25000 + (int32(int16((int16(data[1])<<8)|int16(data[0])))*125)/32
	return
}
