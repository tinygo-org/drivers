// Package lsm303agr implements a driver for the LSM303AGR,
// a 3 axis accelerometer/magnetic sensor which is included on BBC micro:bits v1.5.
//
// Datasheet: https://www.st.com/resource/en/datasheet/lsm303agr.pdf
package lsm303agr // import "tinygo.org/x/drivers/lsm303agr"

import (
	"errors"
	"math"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// Device wraps an I2C connection to a LSM303AGR device.
type Device struct {
	bus            drivers.I2C
	AccelAddress   uint8
	MagAddress     uint8
	AccelPowerMode uint8
	AccelRange     uint8
	AccelDataRate  uint8
	MagPowerMode   uint8
	MagSystemMode  uint8
	MagDataRate    uint8
	buf            [6]uint8
}

// Configuration for LSM303AGR device.
type Configuration struct {
	AccelPowerMode uint8
	AccelRange     uint8
	AccelDataRate  uint8
	MagPowerMode   uint8
	MagSystemMode  uint8
	MagDataRate    uint8
}

var errNotConnected = errors.New("lsm303agr: failed to communicate with either acel or magnet sensor")

// New creates a new LSM303AGR connection. The I2C bus must already be configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) *Device {
	return &Device{
		bus:          bus,
		AccelAddress: ACCEL_ADDRESS,
		MagAddress:   MAG_ADDRESS,
	}
}

// Connected returns whether both sensor on LSM303AGR has been found.
// It does two "who am I" requests and checks the responses.
func (d *Device) Connected() bool {
	data1, data2 := []byte{0}, []byte{0}
	legacy.ReadRegister(d.bus, uint8(d.AccelAddress), ACCEL_WHO_AM_I, data1)
	legacy.ReadRegister(d.bus, uint8(d.MagAddress), MAG_WHO_AM_I, data2)
	return data1[0] == 0x33 && data2[0] == 0x40
}

// Configure sets up the LSM303AGR device for communication.
func (d *Device) Configure(cfg Configuration) (err error) {

	// Verify unit communication
	if !d.Connected() {
		return errNotConnected
	}

	if cfg.AccelDataRate != 0 {
		d.AccelDataRate = cfg.AccelDataRate
	} else {
		d.AccelDataRate = ACCEL_DATARATE_100HZ
	}

	if cfg.AccelPowerMode != 0 {
		d.AccelPowerMode = cfg.AccelPowerMode
	} else {
		d.AccelPowerMode = ACCEL_POWER_NORMAL
	}

	if cfg.AccelRange != 0 {
		d.AccelRange = cfg.AccelRange
	} else {
		d.AccelRange = ACCEL_RANGE_2G
	}

	if cfg.MagPowerMode != 0 {
		d.MagPowerMode = cfg.MagPowerMode
	} else {
		d.MagPowerMode = MAG_POWER_NORMAL
	}

	if cfg.MagDataRate != 0 {
		d.MagDataRate = cfg.MagDataRate
	} else {
		d.MagDataRate = MAG_DATARATE_10HZ
	}

	if cfg.MagSystemMode != 0 {
		d.MagSystemMode = cfg.MagSystemMode
	} else {
		d.MagSystemMode = MAG_SYSTEM_CONTINUOUS
	}

	data := d.buf[:1]

	data[0] = byte(d.AccelDataRate<<4 | d.AccelPowerMode | 0x07)
	err = legacy.WriteRegister(d.bus, uint8(d.AccelAddress), ACCEL_CTRL_REG1_A, data)
	if err != nil {
		return
	}

	data[0] = byte(0x80 | d.AccelRange<<4)
	err = legacy.WriteRegister(d.bus, uint8(d.AccelAddress), ACCEL_CTRL_REG4_A, data)
	if err != nil {
		return
	}

	data[0] = byte(0xC0)
	err = legacy.WriteRegister(d.bus, uint8(d.AccelAddress), TEMP_CFG_REG_A, data)
	if err != nil {
		return
	}

	// Temperature compensation is on for magnetic sensor
	data[0] = byte(0x80 | d.MagPowerMode<<4 | d.MagDataRate<<2 | d.MagSystemMode)
	err = legacy.WriteRegister(d.bus, uint8(d.MagAddress), MAG_MR_REG_M, data)
	if err != nil {
		return
	}

	return nil
}

// ReadAcceleration reads the current acceleration from the device and returns
// it in µg (micro-gravity). When one of the axes is pointing straight to Earth
// and the sensor is not moving the returned value will be around 1000000 or
// -1000000.
func (d *Device) ReadAcceleration() (x, y, z int32, err error) {
	data := d.buf[:6]
	err = legacy.ReadRegister(d.bus, uint8(d.AccelAddress), ACCEL_OUT_AUTO_INC, data)
	if err != nil {
		return
	}

	rangeFactor := int16(0)
	switch d.AccelRange {
	case ACCEL_RANGE_2G:
		rangeFactor = 1
	case ACCEL_RANGE_4G:
		rangeFactor = 2
	case ACCEL_RANGE_8G:
		rangeFactor = 4
	case ACCEL_RANGE_16G:
		rangeFactor = 12 // the readings in 16G are a bit lower
	}

	x = int32(int32(int16((uint16(data[1])<<8|uint16(data[0])))>>4*rangeFactor) * 1000000 / 1024)
	y = int32(int32(int16((uint16(data[3])<<8|uint16(data[2])))>>4*rangeFactor) * 1000000 / 1024)
	z = int32(int32(int16((uint16(data[5])<<8|uint16(data[4])))>>4*rangeFactor) * 1000000 / 1024)
	return
}

// ReadPitchRoll reads the current pitch and roll angles from the device and
// returns it in micro-degrees. When the z axis is pointing straight to Earth
// the returned values of pitch and roll would be zero.
func (d *Device) ReadPitchRoll() (pitch, roll int32, err error) {

	x, y, z, err := d.ReadAcceleration()
	if err != nil {
		return
	}
	xf, yf, zf := float64(x), float64(y), float64(z)
	pitch = int32((math.Round(math.Atan2(yf, math.Sqrt(math.Pow(xf, 2)+math.Pow(zf, 2)))*(180/math.Pi)*100) / 100) * 1000000)
	roll = int32((math.Round(math.Atan2(xf, math.Sqrt(math.Pow(yf, 2)+math.Pow(zf, 2)))*(180/math.Pi)*100) / 100) * 1000000)
	return

}

// ReadMagneticField reads the current magnetic field from the device and returns
// it in mG (milligauss). 1 mG = 0.1 µT (microtesla).
func (d *Device) ReadMagneticField() (x, y, z int32, err error) {

	if d.MagSystemMode == MAG_SYSTEM_SINGLE {
		cmd := d.buf[:1]
		cmd[0] = byte(0x80 | d.MagPowerMode<<4 | d.MagDataRate<<2 | d.MagSystemMode)
		err = legacy.WriteRegister(d.bus, uint8(d.MagAddress), MAG_MR_REG_M, cmd)
		if err != nil {
			return
		}
	}

	data := d.buf[0:6]
	legacy.ReadRegister(d.bus, uint8(d.MagAddress), MAG_OUT_AUTO_INC, data)

	x = int32(int16((uint16(data[1])<<8 | uint16(data[0]))))
	y = int32(int16((uint16(data[3])<<8 | uint16(data[2]))))
	z = int32(int16((uint16(data[5])<<8 | uint16(data[4]))))
	return
}

// ReadCompass reads the current compass heading from the device and returns
// it in micro-degrees. When the z axis is pointing straight to Earth and
// the y axis is pointing to North, the heading would be zero.
//
// However, the heading may be off due to electronic compasses would be effected
// by strong magnetic fields and require constant calibration.
func (d *Device) ReadCompass() (h int32, err error) {

	x, y, _, err := d.ReadMagneticField()
	if err != nil {
		return
	}
	xf, yf := float64(x), float64(y)
	h = int32(float32((180/math.Pi)*math.Atan2(yf, xf)) * 1000000)
	return
}

// ReadTemperature returns the temperature in Celsius milli degrees (°C/1000)
func (d *Device) ReadTemperature() (t int32, err error) {

	data := d.buf[:2]
	err = legacy.ReadRegister(d.bus, uint8(d.AccelAddress), OUT_TEMP_AUTO_INC, data)
	if err != nil {
		return
	}

	r := int16((uint16(data[1])<<8 | uint16(data[0]))) >> 4 // temperature offset from 25 °C
	t = 25000 + int32((float32(r)/8)*1000)
	return
}
