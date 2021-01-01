package bmp388

import (
	"fmt"
	"math"
	"time"

	"tinygo.org/x/drivers"
)

type Oversampling byte
type Mode byte

type Device struct {
	bus         drivers.I2C
	Address     uint8
	cali        calibrationCoefficients
	Pressure    Oversampling
	Temperature Oversampling
	Mode        Mode
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
func (d *Device) Configure(pressure Oversampling, temperature Oversampling, mode Mode) (err error) {
	d.Pressure = pressure
	d.Temperature = temperature
	d.Mode = mode

	// Turning on the pressure and temperature sensors and setting the measurement mode to normal
	if err = d.bus.WriteRegister(d.Address, REG_PWR_CTRL, []byte{PWR_PRESS | PWR_TEMP | byte(mode)}); err != nil {
		return
	}

	// Configure the oversampling settings
	if err = d.bus.WriteRegister(d.Address, REG_OSR, []byte{byte((temperature << 3) | pressure)}); err != nil {
		return
	}

	// Reading the builtin calibration coefficients and parsing them per the datasheet. The compensation formula given
	// in the datasheet is implemented in floating point
	buffer := make([]byte, 21)
	if err = d.bus.ReadRegister(uint8(d.Address), REG_CALI, buffer); err != nil {
		return
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

// ReadTemperature returns the temperature in celsius
func (d *Device) ReadTemperature() (temp float32, err error) {
	buffer := make([]byte, 3)

	// wait until temperature data is ready
	for d.bus.ReadRegister(d.Address, REG_STAT, buffer[0:1]); (buffer[0] & DRDY_TEMP) == 0; d.bus.ReadRegister(d.Address, REG_STAT, buffer[0:1]) {
		time.Sleep(time.Millisecond)
	}

	if err = d.bus.ReadRegister(d.Address, REG_TEMP, buffer); err != nil {
		return 0, fmt.Errorf("[%v] failed to read temperature data", err)
	}

	rawTemp := (float32)(int32(buffer[2])<<16 | int32(buffer[1])<<8 | int32(buffer[0]))

	// eqns are from the compensation formula in the datasheet
	partial1 := rawTemp - d.cali.t1
	partial2 := partial1 * d.cali.t2
	temp = partial2 + (partial1*partial1)*d.cali.t3
	return temp, nil
}

// ReadPressure returns the pressure in pascals
func (d *Device) ReadPressure() (press float32, err error) {
	buffer := make([]byte, 3)

	// wait until pressure data is ready
	for d.bus.ReadRegister(d.Address, REG_STAT, buffer[0:1]); (buffer[0] & DRDY_PRESS) == 0; d.bus.ReadRegister(d.Address, REG_STAT, buffer[0:1]) {
		time.Sleep(time.Millisecond)
	}
	if err = d.bus.ReadRegister(d.Address, REG_PRESS, buffer); err != nil {
		return 0, fmt.Errorf("[%v] failed to read pressure data", err)
	}

	temp, err := d.ReadTemperature()
	if err != nil {
		return
	}

	rawPress := (float32)(int32(buffer[2])<<16 | int32(buffer[1])<<8 | int32(buffer[0]))

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
// the local SEA LEVEL pressure, not the actual pressure.
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

// SoftReset commands the BMP388 to trigger a reset of all user configuration settings
func (d *Device) SoftReset() error {
	err := d.bus.WriteRegister(d.Address, REG_CMD, []byte{SOFT_RESET, 0xB0})
	if err != nil {
		return fmt.Errorf("[%v] failed to perform a soft reset", err)
	}
	return nil
}

// Connected tries to reach the bmp388 and check its chip id register. Returns true if it was able to successfully
// communicate over i2c and returns the correct value
func (d *Device) Connected() bool {
	buffer := make([]byte, 1)
	err := d.bus.ReadRegister(d.Address, REG_CHIP_ID, buffer)
	return err == nil && buffer[0] == CHIP_ID // returns true if i2c comm was good and response equals 0x50
}

func (d *Device) configurationError() bool {
	buffer := make([]byte, 1)
	err := d.bus.ReadRegister(d.Address, REG_ERR, buffer)
	return err == nil && (buffer[0]&0x04) != 0
}
