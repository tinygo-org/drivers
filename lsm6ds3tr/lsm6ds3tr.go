// Package lsm6ds3tr implements a driver for the LSM6DS3TR
// a 6 axis Inertial Measurement Unit (IMU)
//
// Datasheet: https://www.st.com/resource/en/datasheet/lsm6ds3tr.pdf
package lsm6ds3tr // import "tinygo.org/x/drivers/lsm6ds3tr"

import (
	"errors"

	"tinygo.org/x/drivers"
)

type AccelRange uint8
type AccelSampleRate uint8
type AccelBandwidth uint8

type GyroRange uint8
type GyroSampleRate uint8

// Device wraps an I2C connection to a LSM6DS3TR device.
type Device struct {
	bus             drivers.I2C
	Address         uint16
	accelRange      AccelRange
	accelSampleRate AccelSampleRate
	gyroRange       GyroRange
	gyroSampleRate  GyroSampleRate
	buf             [7]uint8 // cmd + up to 6 bytes of data
}

// Configuration for LSM6DS3TR device.
type Configuration struct {
	AccelRange       AccelRange
	AccelSampleRate  AccelSampleRate
	AccelBandWidth   AccelBandwidth
	GyroRange        GyroRange
	GyroSampleRate   GyroSampleRate
	IsPedometer      bool
	ResetStepCounter bool
}

var errNotConnected = errors.New("lsm6ds3tr: failed to communicate with acel/gyro sensor")

// New creates a new LSM6DS3TR connection. The I2C bus must already be configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) *Device {
	return &Device{
		bus:     bus,
		Address: Address,
	}
}

// Configure sets up the device for communication.
func (d *Device) doConfigure(cfg Configuration) (err error) {

	// Verify unit communication
	if !d.Connected() {
		return errNotConnected
	}

	if cfg.AccelRange != 0 {
		d.accelRange = cfg.AccelRange
	} else {
		d.accelRange = ACCEL_2G
	}

	if cfg.AccelSampleRate != 0 {
		d.accelSampleRate = cfg.AccelSampleRate
	} else {
		d.accelSampleRate = ACCEL_SR_104
	}

	if cfg.GyroRange != 0 {
		d.gyroRange = cfg.GyroRange
	} else {
		d.gyroRange = GYRO_2000DPS
	}

	if cfg.GyroSampleRate != 0 {
		d.gyroSampleRate = cfg.GyroSampleRate
	} else {
		d.gyroSampleRate = GYRO_SR_104
	}

	// Configure accelerometer
	d.buf[0] = CTRL1_XL
	d.buf[1] = uint8(d.accelRange) | uint8(d.accelSampleRate)
	err = d.bus.Tx(d.Address, d.buf[0:2], nil)
	if err != nil {
		return
	}

	// Set ODR bit
	d.buf[0] = CTRL4_C
	err = d.bus.Tx(d.Address, d.buf[0:1], d.buf[1:2])
	if err != nil {
		return
	}
	d.buf[0] = CTRL4_C
	d.buf[1] = d.buf[1] &^ BW_SCAL_ODR_ENABLED
	d.buf[1] |= BW_SCAL_ODR_ENABLED
	err = d.bus.Tx(d.Address, d.buf[0:2], nil)
	if err != nil {
		return
	}

	// Configure gyroscope
	d.buf[0] = CTRL2_G
	d.buf[1] = uint8(d.gyroRange) | uint8(d.gyroSampleRate)
	err = d.bus.Tx(d.Address, d.buf[0:2], nil)
	if err != nil {
		return
	}

	return nil
}

// Connected returns whether a LSM6DS3TR has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {
	d.buf[0] = WHO_AM_I
	d.bus.Tx(d.Address, d.buf[0:1], d.buf[1:2])
	return d.buf[1] == 0x6A
}

// ReadAcceleration reads the current acceleration from the device and returns
// it in µg (micro-gravity). When one of the axes is pointing straight to Earth
// and the sensor is not moving the returned value will be around 1000000 or
// -1000000.
func (d *Device) ReadAcceleration() (x, y, z int32, err error) {
	d.buf[0] = OUTX_L_XL
	err = d.bus.Tx(d.Address, d.buf[0:1], d.buf[1:7])
	if err != nil {
		return
	}
	// k comes from "Table 3. Mechanical characteristics" 3 of the datasheet * 1000
	k := int32(61) // 2G
	if d.accelRange == ACCEL_4G {
		k = 122
	} else if d.accelRange == ACCEL_8G {
		k = 244
	} else if d.accelRange == ACCEL_16G {
		k = 488
	}
	x = int32(int16((uint16(d.buf[2])<<8)|uint16(d.buf[1]))) * k
	y = int32(int16((uint16(d.buf[4])<<8)|uint16(d.buf[3]))) * k
	z = int32(int16((uint16(d.buf[6])<<8)|uint16(d.buf[5]))) * k
	return
}

// ReadRotation reads the current rotation from the device and returns it in
// µ°/s (micro-degrees/sec). This means that if you were to do a complete
// rotation along one axis and while doing so integrate all values over time,
// you would get a value close to 360000000.
func (d *Device) ReadRotation() (x, y, z int32, err error) {
	d.buf[0] = OUTX_L_G
	err = d.bus.Tx(d.Address, d.buf[0:1], d.buf[1:7])
	if err != nil {
		return
	}
	// k comes from "Table 3. Mechanical characteristics" 3 of the datasheet * 1000
	k := int32(4375) // 125DPS
	if d.gyroRange == GYRO_245DPS {
		k = 8750
	} else if d.gyroRange == GYRO_500DPS {
		k = 17500
	} else if d.gyroRange == GYRO_1000DPS {
		k = 35000
	} else if d.gyroRange == GYRO_2000DPS {
		k = 70000
	}
	x = int32(int16((uint16(d.buf[2])<<8)|uint16(d.buf[1]))) * k
	y = int32(int16((uint16(d.buf[4])<<8)|uint16(d.buf[3]))) * k
	z = int32(int16((uint16(d.buf[6])<<8)|uint16(d.buf[5]))) * k
	return
}

// ReadTemperature returns the temperature in celsius milli degrees (°C/1000)
func (d *Device) ReadTemperature() (t int32, err error) {
	d.buf[0] = OUT_TEMP_L
	err = d.bus.Tx(d.Address, d.buf[0:1], d.buf[1:3])
	if err != nil {
		return
	}
	// From "Table 5. Temperature sensor characteristics"
	// temp = value/256 + 25
	t = 25000 + (int32(int16((int16(d.buf[2])<<8)|int16(d.buf[1])))*125)/32
	return
}
