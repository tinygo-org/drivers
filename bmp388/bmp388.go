package bmp388

import (
	"errors"
	"math"

	"tinygo.org/x/drivers"
)

type Oversampling byte
type Mode byte
type OutputDataRate byte
type FilterCoefficient byte

type BMP388Config struct {
	Pressure    Oversampling
	Temperature Oversampling
	Mode        Mode
	ODR         OutputDataRate
	IIR         FilterCoefficient
}

type Device struct {
	bus     drivers.I2C
	Address uint8
	cali    calibrationCoefficients
	Config  BMP388Config
}

type calibrationCoefficients struct {
	// Temperature compensation
	t1 float32
	t2 float32
	t3 float32

	// Pressure compensation
	p1  float32
	p2  float32
	p3  float32
	p4  float32
	p5  float32
	p6  float32
	p7  float32
	p8  float32
	p9  float32
	p10 float32
	p11 float32
}

// New returns a bmp388 struct with the default I2C address. Configure must also be called after instanting
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: ADDRESS,
	}
}

// Configure can enable settings on the BMP388 and reads the calibration coefficients
func (d *Device) Configure(config BMP388Config) (err error) {
	d.Config = config

	if d.Config == (BMP388Config{}) {
		d.Config.Mode = NORMAL
	}

	// Turning on the pressure and temperature sensors and setting the measurement mode
	err = d.writeRegister(REG_PWR_CTRL, PWR_PRESS|PWR_TEMP|byte(d.Config.Mode))

	// Configure the oversampling, output data rate, and iir filter coefficient settings
	err = d.writeRegister(REG_OSR, byte(d.Config.Pressure|d.Config.Temperature<<3))
	err = d.writeRegister(REG_ODR, byte(d.Config.ODR))
	err = d.writeRegister(REG_IIR, byte(d.Config.IIR<<1))

	if err != nil {
		return errors.New("bmp388: failed to configure sensor, check connection")
	}

	// Check if there is a problem with the given configuration
	if d.configurationError() {
		return errors.New("bmp388: there is a problem with the configuration, try reducing ODR")
	}

	// Reading the builtin calibration coefficients and parsing them per the datasheet. The compensation formula given
	// in the datasheet is implemented in floating point
	buffer, err := d.readRegister(REG_CALI, 21)
	if err != nil {
		return errors.New("bmp388: failed to read calibration coefficient register")
	}

	t1 := uint16(buffer[1])<<8 | uint16(buffer[0])
	t2 := uint16(buffer[3])<<8 | uint16(buffer[2])
	t3 := int8(buffer[4])

	p1 := int16(buffer[6])<<8 | int16(buffer[5])
	p2 := int16(buffer[8])<<8 | int16(buffer[7])
	p3 := int8(buffer[9])
	p4 := int8(buffer[10])
	p5 := uint16(buffer[12])<<8 | uint16(buffer[11])
	p6 := uint16(buffer[14])<<8 | uint16(buffer[13])
	p7 := int8(buffer[15])
	p8 := int8(buffer[16])
	p9 := int16(buffer[18])<<8 | int16(buffer[17])
	p10 := int8(buffer[19])
	p11 := int8(buffer[20])

	d.cali.t1 = float32(t1) * float32(1<<8)
	d.cali.t2 = float32(t2) / float32(1<<30)
	d.cali.t3 = float32(t3) / float32(1<<48)

	d.cali.p1 = (float32(p1) - float32(1<<14)) / float32(1<<20)
	d.cali.p2 = (float32(p2) - float32(1<<14)) / float32(1<<29)
	d.cali.p3 = float32(p3) / float32(1<<32)
	d.cali.p4 = float32(p4) / float32(1<<37)
	d.cali.p5 = float32(p5) * float32(1<<3)
	d.cali.p6 = float32(p6) / float32(1<<6)
	d.cali.p7 = float32(p7) / float32(1<<8)
	d.cali.p8 = float32(p8) / float32(1<<15)
	d.cali.p9 = float32(p9) / float32(1<<48)
	d.cali.p10 = float32(p10) / float32(1<<48)
	d.cali.p11 = float32(p11) / float32(1<<65)

	return nil
}

func (d *Device) readSensorData(register byte) (data float32, err error) {

	// put the sensor back into forced mode to get a reading, the sensor goes back to sleep after taking one read in
	// forced mode
	if d.Config.Mode != NORMAL {
		err = d.SetMode(FORCED)
		if err != nil {
			return
		}
	}

	bytes, err := d.readRegister(register, 3)
	if err != nil {
		return
	}
	data = (float32)(int32(bytes[2])<<16 | int32(bytes[1])<<8 | int32(bytes[0]))
	return
}

// ReadTemperature returns the temperature in celsius
func (d *Device) ReadTemperature() (temp float32, err error) {

	rawTemp, err := d.readSensorData(REG_TEMP)
	if err != nil {
		return
	}

	// eqns are from the compensation formula in the datasheet
	partial1 := rawTemp - d.cali.t1
	partial2 := partial1 * d.cali.t2
	temp = partial2 + (partial1*partial1)*d.cali.t3
	return temp, nil
}

// ReadPressure returns the pressure in pascals
func (d *Device) ReadPressure() (press float32, err error) {

	temp, err := d.ReadTemperature()
	if err != nil {
		return
	}
	rawPress, err := d.readSensorData(REG_PRESS)
	if err != nil {
		return
	}

	partial1 := d.cali.p6 * temp
	partial2 := d.cali.p7 * (temp * temp)
	partial3 := d.cali.p8 * (temp * temp * temp)
	partialOut1 := d.cali.p5 + partial1 + partial2 + partial3

	partial1 = d.cali.p2 * temp
	partial2 = d.cali.p3 * (temp * temp)
	partial3 = d.cali.p4 * (temp * temp * temp)
	partialOut2 := rawPress * (d.cali.p1 + partial1 + partial2 + partial3)

	partial1 = rawPress * rawPress
	partial2 = d.cali.p9 + d.cali.p10*temp
	partial3 = partial1 * partial2
	partialOut3 := partial3 + (rawPress*rawPress*rawPress)*d.cali.p11
	press = partialOut1 + partialOut2 + partialOut3

	return press, nil
}

// ReadAltitude estimates the altitude above sea level in meters. The equation is only valid below 11 km. refPress is
// the LOCAL SEA LEVEL pressure, not the actual pressure.
func (d *Device) ReadAltitude(refPress float32) (alt float32, err error) {
	press, err := d.ReadPressure()
	temp, err := d.ReadTemperature()
	if err != nil {
		return
	}

	// This equation is only valid below 11 km
	alt = (float32(math.Pow(float64(refPress)/float64(press), (1/5.257))-1) * (temp + 273.15)) / 0.0065
	return alt, nil
}

// SoftReset commands the BMP388 to reset of all user configuration settings
func (d *Device) SoftReset() error {
	err := d.writeRegister(REG_CMD, SOFT_RESET)
	if err != nil {
		return errors.New("bmp388: failed to perform a soft reset")
	}
	return nil
}

// Connected tries to reach the bmp388 and check its chip id register. Returns true if it was able to successfully
// communicate over i2c and returns the correct value
func (d *Device) Connected() bool {
	data, err := d.readRegister(REG_CHIP_ID, 1)
	return err == nil && data[0] == CHIP_ID // returns true if i2c comm was good and response equals 0x50
}

// ConfigurationError checks the register error for the configuration error bit. The bit is cleared on read by the bmp.
func (d *Device) configurationError() bool {
	data, err := d.readRegister(REG_ERR, 1)
	return err == nil && (data[0]&0x04) != 0
}

// SetMode changes the run mode of the sensor, NORMAL is the one to use for most cases. Use FORCED if you plan to take
// measurements infrequently and want to conserve power. SLEEP will of course put the sensor to sleep
func (d *Device) SetMode(mode Mode) error {
	d.Config.Mode = mode
	return d.writeRegister(REG_PWR_CTRL, PWR_PRESS|PWR_TEMP|byte(d.Config.Mode))
}

func (d *Device) readRegister(register byte, len int) (data []byte, err error) {
	data = make([]byte, len)
	err = d.bus.ReadRegister(d.Address, register, data)
	return
}

func (d *Device) writeRegister(register byte, data byte) error {
	return d.bus.WriteRegister(d.Address, register, []byte{data})
}
