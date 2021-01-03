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
	Pressure         Oversampling
	Temperature      Oversampling
	Mode             Mode
	ODR              OutputDataRate
	IIR              FilterCoefficient
	SeaLevelPressure int32 // sea level pressure in centipascals
}

type Device struct {
	bus     drivers.I2C
	Address uint8
	cali    calibrationCoefficients
	Config  BMP388Config
}

type calibrationCoefficients struct {
	// Temperature compensation
	t1 uint16
	t2 uint16
	t3 int8

	// Pressure compensation
	p1  int16
	p2  int16
	p3  int8
	p4  int8
	p5  uint16
	p6  uint16
	p7  int8
	p8  int8
	p9  int16
	p10 int8
	p11 int8

	// temperature value used for pressure compensation
	tlin int64
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

	if d.Config.SeaLevelPressure == 0 {
		d.Config.SeaLevelPressure = 10132500 // standard sea level pressure in cPa
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

	d.cali.t1 = uint16(buffer[1])<<8 | uint16(buffer[0])
	d.cali.t2 = uint16(buffer[3])<<8 | uint16(buffer[2])
	d.cali.t3 = int8(buffer[4])

	d.cali.p1 = int16(buffer[6])<<8 | int16(buffer[5])
	d.cali.p2 = int16(buffer[8])<<8 | int16(buffer[7])
	d.cali.p3 = int8(buffer[9])
	d.cali.p4 = int8(buffer[10])
	d.cali.p5 = uint16(buffer[12])<<8 | uint16(buffer[11])
	d.cali.p6 = uint16(buffer[14])<<8 | uint16(buffer[13])
	d.cali.p7 = int8(buffer[15])
	d.cali.p8 = int8(buffer[16])
	d.cali.p9 = int16(buffer[18])<<8 | int16(buffer[17])
	d.cali.p10 = int8(buffer[19])
	d.cali.p11 = int8(buffer[20])

	return nil
}

// ReadTemperature returns the temperature in centicelsius, i.e 2426 / 100 = 24.26 C
func (d *Device) ReadTemperature() (int32, error) {

	rawTemp, err := d.readSensorData(REG_TEMP)
	if err != nil {
		return 0, err
	}

	// pulled from C driver: https://github.com/BoschSensortec/BMP3-Sensor-API/blob/master/bmp3.c
	partialData1 := rawTemp - (256 * int64(d.cali.t1))
	partialData2 := int64(d.cali.t2) * partialData1
	partialData3 := (partialData1 * partialData1)
	partialData4 := partialData3 * int64(d.cali.t3)
	partialData5 := (partialData2 * 262144) + partialData4
	partialData6 := partialData5 / 4294967296

	d.cali.tlin = partialData6
	temp := ((partialData6 * 25) / 16384)
	return int32(temp), nil
}

// ReadPressure returns the pressure in centipascals, i.e 10132520 / 100 = 101325.20 Pa
func (d *Device) ReadPressure() (int32, error) {

	_, err := d.ReadTemperature()
	if err != nil {
		return 0, err
	}
	rawPress, err := d.readSensorData(REG_PRESS)
	if err != nil {
		return 0, err
	}

	// code pulled from bmp388 C driver: https://github.com/BoschSensortec/BMP3-Sensor-API/blob/master/bmp3.c
	partialData1 := d.cali.tlin * d.cali.tlin
	partialData2 := partialData1 / 64
	partialData3 := (partialData2 * d.cali.tlin) / 256
	partialData4 := (int64(d.cali.p8) * partialData3) / 32
	partialData5 := (int64(d.cali.p7) * partialData1) * 16
	partialData6 := (int64(d.cali.p6) * d.cali.tlin) * 4194304
	offset := (int64(d.cali.p5) * 140737488355328) + partialData4 + partialData5 + partialData6
	partialData2 = (int64(d.cali.p4) * partialData3) / 32
	partialData4 = (int64(d.cali.p3) * partialData1) * 4
	partialData5 = (int64(d.cali.p2) - 16384) * d.cali.tlin * 2097152
	sensitivity := ((int64(d.cali.p1) - 16384) * 70368744177664) + partialData2 + partialData4 + partialData5
	partialData1 = (sensitivity / 16777216) * rawPress
	partialData2 = int64(d.cali.p10) * d.cali.tlin
	partialData3 = partialData2 + (65536 * int64(d.cali.p9))
	partialData4 = (partialData3 * rawPress) / 8192

	// dividing by 10 followed by multiplying by 10
	// To avoid overflow caused by (pressure * partial_data4)
	partialData5 = (rawPress * (partialData4 / 10)) / 512
	partialData5 = partialData5 * 10
	partialData6 = (int64)(uint64(rawPress) * uint64(rawPress))
	partialData2 = (int64(d.cali.p11) * partialData6) / 65536
	partialData3 = (partialData2 * rawPress) / 128
	partialData4 = (offset / 4) + partialData1 + partialData5 + partialData3
	compPress := ((uint64(partialData4) * 25) / uint64(1099511627776))
	return int32(compPress), nil
}

// ReadAltitude estimates the altitude above sea level in centimeters.
func (d *Device) ReadAltitude() (alt int32, err error) {
	press, err := d.ReadPressure()
	if err != nil {
		return
	}
	// pulled from bme280 driver, modified a bit to return altitude in centimeters
	alt = int32(4433000.0 * (1.0 - math.Pow(float64(press)/float64(d.Config.SeaLevelPressure), 0.1903)))
	return
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

// SetMode changes the run mode of the sensor, NORMAL is the one to use for most cases. Use FORCED if you plan to take
// measurements infrequently and want to conserve power. SLEEP will of course put the sensor to sleep
func (d *Device) SetMode(mode Mode) error {
	d.Config.Mode = mode
	return d.writeRegister(REG_PWR_CTRL, PWR_PRESS|PWR_TEMP|byte(d.Config.Mode))
}

func (d *Device) readSensorData(register byte) (data int64, err error) {

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
	data = int64(bytes[2])<<16 | int64(bytes[1])<<8 | int64(bytes[0])
	return
}

// configurationError checks the register error for the configuration error bit. The bit is cleared on read by the bmp.
func (d *Device) configurationError() bool {
	data, err := d.readRegister(REG_ERR, 1)
	return err == nil && (data[0]&0x04) != 0
}

func (d *Device) readRegister(register byte, len int) (data []byte, err error) {
	data = make([]byte, len)
	err = d.bus.ReadRegister(d.Address, register, data)
	return
}

func (d *Device) writeRegister(register byte, data byte) error {
	return d.bus.WriteRegister(d.Address, register, []byte{data})
}
