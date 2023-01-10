// Package bme280 provides a driver for the BME280 digital combined
// humidity and pressure sensor by Bosch.
//
// Datasheet:
// https://cdn-shop.adafruit.com/datasheets/BST-BME280_DS001-10.pdf
package bme280

import (
	"math"
	"time"

	"tinygo.org/x/drivers"
)

// calibrationCoefficients reads at startup and stores the calibration coefficients
type calibrationCoefficients struct {
	t1 uint16
	t2 int16
	t3 int16
	p1 uint16
	p2 int16
	p3 int16
	p4 int16
	p5 int16
	p6 int16
	p7 int16
	p8 int16
	p9 int16
	h1 uint8
	h2 int16
	h3 uint8
	h4 int16
	h5 int16
	h6 int8
}

type Oversampling byte
type Mode byte
type FilterCoefficient byte
type Period byte

// Config contains settings for filtering, sampling, and modes of operation
type Config struct {
	Pressure    Oversampling
	Temperature Oversampling
	Humidity    Oversampling
	Period      Period
	Mode        Mode
	IIR         FilterCoefficient
}

// Device wraps an I2C connection to a BME280 device.
type Device struct {
	bus                     drivers.I2C
	Address                 uint16
	calibrationCoefficients calibrationCoefficients
	Config                  Config
}

// New creates a new BME280 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
	}
}

// ConfigureWithSettings sets up the device for communication and
// read the calibration coefficients.
//
// The default configuration is the Indoor Navigation settings
// from the BME280 datasheet.
func (d *Device) Configure() {
	d.ConfigureWithSettings(Config{})
}

// ConfigureWithSettings sets up the device for communication and
// read the calibration coefficients.
//
// The default configuration if config is left at defaults is
// the Indoor Navigation settings from the BME280 datasheet.
func (d *Device) ConfigureWithSettings(config Config) {
	d.Config = config

	// If config is not initialized, use Indoor Navigation defaults.
	if d.Config == (Config{}) {
		d.Config = Config{
			Mode:        ModeNormal,
			Period:      Period0_5ms,
			Temperature: Sampling2X,
			Humidity:    Sampling1X,
			Pressure:    Sampling16X,
			IIR:         Coeff16,
		}
	}

	var data [24]byte
	err := d.bus.ReadRegister(uint8(d.Address), REG_CALIBRATION, data[:])
	if err != nil {
		return
	}

	var h1 [1]byte
	err = d.bus.ReadRegister(uint8(d.Address), REG_CALIBRATION_H1, h1[:])
	if err != nil {
		return
	}

	var h2lsb [7]byte
	err = d.bus.ReadRegister(uint8(d.Address), REG_CALIBRATION_H2LSB, h2lsb[:])
	if err != nil {
		return
	}

	d.calibrationCoefficients.t1 = readUintLE(data[0], data[1])
	d.calibrationCoefficients.t2 = readIntLE(data[2], data[3])
	d.calibrationCoefficients.t3 = readIntLE(data[4], data[5])
	d.calibrationCoefficients.p1 = readUintLE(data[6], data[7])
	d.calibrationCoefficients.p2 = readIntLE(data[8], data[9])
	d.calibrationCoefficients.p3 = readIntLE(data[10], data[11])
	d.calibrationCoefficients.p4 = readIntLE(data[12], data[13])
	d.calibrationCoefficients.p5 = readIntLE(data[14], data[15])
	d.calibrationCoefficients.p6 = readIntLE(data[16], data[17])
	d.calibrationCoefficients.p7 = readIntLE(data[18], data[19])
	d.calibrationCoefficients.p8 = readIntLE(data[20], data[21])
	d.calibrationCoefficients.p9 = readIntLE(data[22], data[23])

	d.calibrationCoefficients.h1 = h1[0]
	d.calibrationCoefficients.h2 = readIntLE(h2lsb[0], h2lsb[1])
	d.calibrationCoefficients.h3 = h2lsb[2]
	d.calibrationCoefficients.h6 = int8(h2lsb[6])
	d.calibrationCoefficients.h4 = 0 + (int16(h2lsb[3]) << 4) | (int16(h2lsb[4] & 0x0F))
	d.calibrationCoefficients.h5 = 0 + (int16(h2lsb[5]) << 4) | (int16(h2lsb[4]) >> 4)

	d.Reset()

	d.bus.WriteRegister(uint8(d.Address), CTRL_CONFIG, []byte{byte(d.Config.Period<<5) | byte(d.Config.IIR<<2)})
	d.bus.WriteRegister(uint8(d.Address), CTRL_HUMIDITY_ADDR, []byte{byte(d.Config.Humidity)})

	// Normal mode, start measuring now
	if d.Config.Mode == ModeNormal {
		d.bus.WriteRegister(uint8(d.Address), CTRL_MEAS_ADDR, []byte{
			byte(d.Config.Temperature<<5) |
				byte(d.Config.Pressure<<2) |
				byte(d.Config.Mode)})
	}
}

// Connected returns whether a BME280 has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {
	data := []byte{0}
	d.bus.ReadRegister(uint8(d.Address), WHO_AM_I, data)
	return data[0] == CHIP_ID
}

// Reset the device
func (d *Device) Reset() {
	d.bus.WriteRegister(uint8(d.Address), CMD_RESET, []byte{0xB6})
}

// SetMode can set the device to Sleep, Normal or Forced mode
//
// Calling this method is optional, Configure can be used to set the
// initial mode if no mode change is desired.  This method is most
// useful to switch between Sleep and Normal modes.
func (d *Device) SetMode(mode Mode) {
	d.Config.Mode = mode

	d.bus.WriteRegister(uint8(d.Address), CTRL_MEAS_ADDR, []byte{
		byte(d.Config.Temperature<<5) |
			byte(d.Config.Pressure<<2) |
			byte(d.Config.Mode)})
}

// ReadTemperature returns the temperature in celsius milli degrees (°C/1000)
func (d *Device) ReadTemperature() (int32, error) {
	data, err := d.readData()
	if err != nil {
		return 0, err
	}

	temp, _ := d.calculateTemp(data)
	return temp, nil
}

// ReadPressure returns the pressure in milli pascals mPa
func (d *Device) ReadPressure() (int32, error) {
	data, err := d.readData()
	if err != nil {
		return 0, err
	}
	_, tFine := d.calculateTemp(data)
	pressure := d.calculatePressure(data, tFine)
	return pressure, nil
}

// ReadHumidity returns the relative humidity in hundredths of a percent
func (d *Device) ReadHumidity() (int32, error) {
	data, err := d.readData()
	if err != nil {
		return 0, err
	}
	_, tFine := d.calculateTemp(data)
	humidity := d.calculateHumidity(data, tFine)
	return humidity, nil
}

// ReadAltitude returns the current altitude in meters based on the
// current barometric pressure and estimated pressure at sea level.
// Calculation is based on code from Adafruit BME280 library
//
//	https://github.com/adafruit/Adafruit_BME280_Library
func (d *Device) ReadAltitude() (alt int32, err error) {
	mPa, _ := d.ReadPressure()
	atmP := float32(mPa) / 100000
	alt = int32(44330.0 * (1.0 - math.Pow(float64(atmP/SEALEVEL_PRESSURE), 0.1903)))
	return
}

// convert2Bytes converts two bytes to int32
func convert2Bytes(msb byte, lsb byte) int32 {
	return int32(readUint(msb, lsb))
}

// convert3Bytes converts three bytes to int32
func convert3Bytes(msb byte, b1 byte, lsb byte) int32 {
	return int32(((((uint32(msb) << 8) | uint32(b1)) << 8) | uint32(lsb)) >> 4)
}

// readUint converts two bytes to uint16
func readUint(msb byte, lsb byte) uint16 {
	return (uint16(msb) << 8) | uint16(lsb)
}

// readUintLE converts two little endian bytes to uint16
func readUintLE(msb byte, lsb byte) uint16 {
	temp := readUint(msb, lsb)
	return (temp >> 8) | (temp << 8)
}

// readIntLE converts two little endian bytes to int16
func readIntLE(msb byte, lsb byte) int16 {
	return int16(readUintLE(msb, lsb))
}

// readData does a burst read from 0xF7 to 0xF0 according to the datasheet
// resulting in an slice with 8 bytes 0-2 = pressure / 3-5 = temperature / 6-7 = humidity
func (d *Device) readData() (data [8]byte, err error) {
	if d.Config.Mode == ModeForced {
		// Write the CTRL_MEAS register to trigger a measurement
		d.bus.WriteRegister(uint8(d.Address), CTRL_MEAS_ADDR, []byte{
			byte(d.Config.Temperature<<5) |
				byte(d.Config.Pressure<<2) |
				byte(d.Config.Mode)})

		time.Sleep(d.measurementDelay())
	}

	err = d.bus.ReadRegister(uint8(d.Address), REG_PRESSURE, data[:])
	if err != nil {
		println(err)
		return
	}
	return
}

// calculateTemp uses the data slice and applies calibrations values on it to convert the value to milli degrees
// it also calculates the variable tFine which is used by the pressure and humidity calculation
func (d *Device) calculateTemp(data [8]byte) (int32, int32) {

	rawTemp := convert3Bytes(data[3], data[4], data[5])

	var1 := (((rawTemp >> 3) - (int32(d.calibrationCoefficients.t1) << 1)) * int32(d.calibrationCoefficients.t2)) >> 11
	var2 := (((((rawTemp >> 4) - int32(d.calibrationCoefficients.t1)) * ((rawTemp >> 4) - int32(d.calibrationCoefficients.t1))) >> 12) * int32(d.calibrationCoefficients.t3)) >> 14

	tFine := var1 + var2
	T := (tFine*5 + 128) >> 8
	return (10 * T), tFine
}

// calculatePressure uses the data slice and applies calibrations values on it to convert the value to milli pascals mPa
func (d *Device) calculatePressure(data [8]byte, tFine int32) int32 {

	rawPressure := convert3Bytes(data[0], data[1], data[2])

	var1 := int64(tFine) - 128000
	var2 := var1 * var1 * int64(d.calibrationCoefficients.p6)
	var2 = var2 + ((var1 * int64(d.calibrationCoefficients.p5)) << 17)
	var2 = var2 + (int64(d.calibrationCoefficients.p4) << 35)
	var1 = ((var1 * var1 * int64(d.calibrationCoefficients.p3)) >> 8) + ((var1 * int64(d.calibrationCoefficients.p2)) << 12)
	var1 = ((int64(1) << 47) + var1) * int64(d.calibrationCoefficients.p1) >> 33

	if var1 == 0 {
		return 0 // avoid exception caused by division by zero
	}
	p := int64(1048576 - rawPressure)
	p = (((p << 31) - var2) * 3125) / var1
	var1 = (int64(d.calibrationCoefficients.p9) * (p >> 13) * (p >> 13)) >> 25
	var2 = (int64(d.calibrationCoefficients.p8) * p) >> 19

	p = ((p + var1 + var2) >> 8) + (int64(d.calibrationCoefficients.p7) << 4)
	p = (p / 256)
	return int32(1000 * p)
}

// calculateHumidity uses the data slice and applies calibrations values on it to convert the value to relative humidity in hundredths of a percent
func (d *Device) calculateHumidity(data [8]byte, tFine int32) int32 {

	rawHumidity := convert2Bytes(data[6], data[7])

	h := float32(tFine) - 76800

	if h == 0 {
		println("invalid value")
	}

	var1 := float32(rawHumidity) - (float32(d.calibrationCoefficients.h4)*64.0 +
		(float32(d.calibrationCoefficients.h5) / 16384.0 * h))

	var2 := float32(d.calibrationCoefficients.h2) / 65536.0 *
		(1.0 + float32(d.calibrationCoefficients.h6)/67108864.0*h*
			(1.0+float32(d.calibrationCoefficients.h3)/67108864.0*h))

	h = var1 * var2
	h = h * (1 - float32(d.calibrationCoefficients.h1)*h/524288)
	return int32(100 * h)

}

// measurementDelay returns how much time each measurement will take
// on the device.
//
// This is used in forced mode to wait until a measurement is complete.
func (d *Device) measurementDelay() time.Duration {
	const MeasOffset = 1250
	const MeasDur = 2300
	const HumMeasOffset = 575
	const MeasScalingFactor = 1000

	// delay is based on over-sampling rate - this table converts from
	// setting to number samples
	sampleRateConv := []int{0, 1, 2, 4, 8, 16}

	tempOsr := 16
	if d.Config.Temperature <= Sampling16X {
		tempOsr = sampleRateConv[d.Config.Temperature]
	}

	presOsr := 16
	if d.Config.Temperature <= Sampling16X {
		presOsr = sampleRateConv[d.Config.Pressure]
	}

	humOsr := 16
	if d.Config.Temperature <= Sampling16X {
		humOsr = sampleRateConv[d.Config.Humidity]
	}

	max_delay := ((MeasOffset + (MeasDur * tempOsr) +
		((MeasDur * presOsr) + HumMeasOffset) +
		((MeasDur * humOsr) + HumMeasOffset)) / MeasScalingFactor)

	return time.Duration(max_delay) * time.Millisecond
}
