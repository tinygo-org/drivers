package bmp388

import (
	"math"

	"tinygo.org/x/drivers"
)

// OversamplingMode is the oversampling ratio of the temperature or pressure measurement.
type Oversampling uint

// Mode is the Power Mode.
type Mode uint

// Standby is the inactive period between the reads when the sensor is in normal power mode.
type Standby uint

// Filter unwanted changes in measurement caused by external (environmental) or internal changes (IC).
type Filter uint

type Device struct {
	bus      drivers.I2C
	Address  uint8
	cali     calibrationCoefficients
	refPress float32
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

func New(bus drivers.I2C, refPress float32) Device {
	return Device{
		bus:      bus,
		Address:  Address,
		refPress: refPress,
	}
}

// Configure can enable settings on the BMP388 and reads the calibration coefficients. The coefficients are converted to
// their floating point counterparts from the equations given in the datasheet
func (d *Device) Configure() error {
	buffer := make([]byte, 21)

	err := d.bus.ReadRegister(uint8(d.Address), REG_CALI, buffer)
	if err != nil {
		return err
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

	err = d.bus.ReadRegister(d.Address, REG_TEMP, buffer)
	if err != nil {
		return
	}

	rawTemp := (float32)(int32(buffer[2])<<16 | int32(buffer[1])<<8 | int32(buffer[0]))

	partial1 := rawTemp - d.cali.t1
	partial2 := partial1 * d.cali.t2
	temp = partial2 + (partial1*partial1)*d.cali.t3
	return temp, nil
}

// ReadPressure returns the pressure in pascals
func (d *Device) ReadPressure() (press float32, err error) {
	buffer := make([]byte, 3)
	err = d.bus.ReadRegister(d.Address, REG_PRESS, buffer)
	if err != nil {
		return
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

// ReadAltitude predicts the altitude above sea level in meters, by using the reference pressure in the bmp388 struct
func (d *Device) ReadAltitude() (alt float32, err error) {
	press, err := d.ReadPressure()
	temp, err := d.ReadTemperature()
	if err != nil {
		return
	}
	alt = (float32(math.Pow(float64(d.refPress)/float64(press), (1/5.257))-1) * (temp + 273.15)) / 0.0065
	return alt, nil
}
