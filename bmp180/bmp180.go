// Package bmp180 provides a driver for the BMP180 digital pressure sensor
// by Bosch.
//
// Datasheet:
// https://cdn-shop.adafruit.com/datasheets/BST-BMP180-DS000-09.pdf
package bmp180 // import "tinygo.org/x/drivers/bmp180"

import (
	"math"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// OversamplingMode is the oversampling ratio of the pressure measurement.
type OversamplingMode uint

// calibrationCoefficients reads at startup and stores the calibration coefficients
type calibrationCoefficients struct {
	ac1 int16
	ac2 int16
	ac3 int16
	ac4 uint16
	ac5 uint16
	ac6 uint16
	b1  int16
	b2  int16
	mb  int16
	mc  int16
	md  int16
}

// Device wraps an I2C connection to a BMP180 device.
type Device struct {
	bus                     drivers.I2C
	Address                 uint16
	mode                    OversamplingMode
	calibrationCoefficients calibrationCoefficients
}

// New creates a new BMP180 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not initialize the device.
// You must call Configure() first in order to use the device itself.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
		mode:    ULTRAHIGHRESOLUTION,
	}
}

// Connected returns whether a BMP180 has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {
	data := []byte{0}
	legacy.ReadRegister(d.bus, uint8(d.Address), WHO_AM_I, data)
	return data[0] == CHIP_ID
}

// Configure sets up the device for communication and
// read the calibration coefficients.
func (d *Device) Configure() {
	data := make([]byte, 22)
	err := legacy.ReadRegister(d.bus, uint8(d.Address), AC1_MSB, data)
	if err != nil {
		return
	}
	d.calibrationCoefficients.ac1 = readInt(data[0], data[1])
	d.calibrationCoefficients.ac2 = readInt(data[2], data[3])
	d.calibrationCoefficients.ac3 = readInt(data[4], data[5])
	d.calibrationCoefficients.ac4 = readUint(data[6], data[7])
	d.calibrationCoefficients.ac5 = readUint(data[8], data[9])
	d.calibrationCoefficients.ac6 = readUint(data[10], data[11])
	d.calibrationCoefficients.b1 = readInt(data[12], data[13])
	d.calibrationCoefficients.b2 = readInt(data[14], data[15])
	d.calibrationCoefficients.mb = readInt(data[16], data[17])
	d.calibrationCoefficients.mc = readInt(data[18], data[19])
	d.calibrationCoefficients.md = readInt(data[20], data[21])
}

// ReadTemperature returns the temperature in celsius milli degrees (°C/1000).
func (d *Device) ReadTemperature() (temperature int32, err error) {
	rawTemp, err := d.rawTemp()
	if err != nil {
		return
	}
	b5 := d.calculateB5(rawTemp)
	t := (b5 + 8) >> 4
	return 100 * t, nil
}

// ReadPressure returns the pressure in milli pascals (mPa).
func (d *Device) ReadPressure() (pressure int32, err error) {
	rawTemp, err := d.rawTemp()
	if err != nil {
		return
	}
	rawPressure, err := d.rawPressure(d.mode)
	if err != nil {
		return
	}
	b5 := d.calculateB5(rawTemp)
	b6 := b5 - 4000
	x1 := (int32(d.calibrationCoefficients.b2) * (b6 * b6 >> 12)) >> 11
	x2 := (int32(d.calibrationCoefficients.ac2) * b6) >> 11
	x3 := x1 + x2
	b3 := (((int32(d.calibrationCoefficients.ac1)*4 + x3) << uint(d.mode)) + 2) >> 2
	x1 = (int32(d.calibrationCoefficients.ac3) * b6) >> 13
	x2 = (int32(d.calibrationCoefficients.b1) * ((b6 * b6) >> 12)) >> 16
	x3 = ((x1 + x2) + 2) >> 2
	b4 := (uint32(d.calibrationCoefficients.ac4) * uint32(x3+32768)) >> 15
	b7 := uint32(rawPressure-b3) * (50000 >> uint(d.mode))
	var p int32
	if b7 < 0x80000000 {
		p = int32((b7 << 1) / b4)
	} else {
		p = int32((b7 / b4) << 1)
	}
	x1 = (p >> 8) * (p >> 8)
	x1 = (x1 * 3038) >> 16
	x2 = (-7357 * p) >> 16
	return 1000 * (p + ((x1 + x2 + 3791) >> 4)), nil
}

// ReadAltitude returns the current altitude in meters based on the
// current barometric pressure and estimated pressure at sea level.
// Calculation is based on code from Adafruit BME280 library
//
// https://github.com/adafruit/Adafruit_BME280_Library
func (d *Device) ReadAltitude() (int32, error) {
	mPa, err := d.ReadPressure()
	if err != nil {
		return 0, err
	}
	atmP := float32(mPa) / 100000

	return int32(44330.0 * (1.0 - math.Pow(float64(atmP/SEALEVEL_PRESSURE), 0.1903))), nil
}

// rawTemp returns the sensor's raw values of the temperature
func (d *Device) rawTemp() (int32, error) {
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_CTRL, []byte{CMD_TEMP})
	time.Sleep(5 * time.Millisecond)
	data := make([]byte, 2)
	err := legacy.ReadRegister(d.bus, uint8(d.Address), REG_TEMP_MSB, data)
	if err != nil {
		return 0, err
	}
	return int32(uint16(data[0])<<8 | uint16(data[1])), nil
}

// calculateB5 calculates intermediate value B5 as per page 15 of datasheet
func (d *Device) calculateB5(rawTemp int32) int32 {
	x1 := (rawTemp - int32(d.calibrationCoefficients.ac6)) * int32(d.calibrationCoefficients.ac5) >> 15
	x2 := int32(d.calibrationCoefficients.mc) << 11 / (x1 + int32(d.calibrationCoefficients.md))
	return x1 + x2
}

// rawPressure returns the sensor's raw values of the pressure
func (d *Device) rawPressure(mode OversamplingMode) (int32, error) {
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_CTRL, []byte{CMD_PRESSURE + byte(mode<<6)})
	time.Sleep(pauseForReading(mode))
	data := make([]byte, 3)
	err := legacy.ReadRegister(d.bus, uint8(d.Address), REG_PRESSURE_MSB, data)
	if err != nil {
		return 0, err
	}
	rawPressure := int32((uint32(data[0])<<16 + uint32(data[1])<<8 + uint32(data[2])) >> (8 - uint(mode)))
	return rawPressure, nil
}

// pauseForReading returns the pause duration depending on the sampling mode
func pauseForReading(mode OversamplingMode) time.Duration {
	var d time.Duration
	switch mode {
	case ULTRALOWPOWER:
		d = 5 * time.Millisecond
	case STANDARD:
		d = 8 * time.Millisecond
	case HIGHRESOLUTION:
		d = 14 * time.Millisecond
	case ULTRAHIGHRESOLUTION:
		d = 26 * time.Millisecond
	}
	return d
}

// readInt converts two bytes to int16
func readInt(msb byte, lsb byte) int16 {
	return int16(uint16(msb)<<8 | uint16(lsb))
}

// readUint converts two bytes to uint16
func readUint(msb byte, lsb byte) uint16 {
	return (uint16(msb) << 8) | uint16(lsb)
}
