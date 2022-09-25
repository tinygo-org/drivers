// Package scd4x provides a driver for the scd4x I2C envrironment sensor.
//
// Datasheet: https://sensirion.com/media/documents/C4B87CE6/627C2DCD/CD_DS_SCD40_SCD41_Datasheet_D1.pdf
//
// This driver is heavily influenced by the scd4x code from Adafruit for CircuitPython:
// https://github.com/adafruit/Adafruit_CircuitPython_SCD4X
// Thank you!
package scd4x // import "tinygo.org/x/drivers/scd4x"

import (
	"encoding/binary"
	"time"

	"tinygo.org/x/drivers"
)

type Device struct {
	bus     drivers.I2C
	tx      []byte
	rx      []byte
	Address uint8

	// used to cache the most recent readings
	co2         uint16
	temperature uint16
	humidity    uint16
}

// New returns SCD4x device for the provided I2C bus using default address of 0x62.
func New(i2c drivers.I2C) *Device {
	return &Device{
		bus:     i2c,
		tx:      make([]byte, 5),
		rx:      make([]byte, 18),
		Address: Address,
	}
}

// Configure the device.
func (d *Device) Configure() (err error) {
	if err := d.StopPeriodicMeasurement(); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)

	// reset the chip
	if err := d.sendCommand(CmdReinit); err != nil {
		return err
	}

	time.Sleep(20 * time.Millisecond)
	return
}

// Connected returns whether sensor has been found.
func (d *Device) Connected() bool {
	// TODO: something here to check if the sensor is connected
	return true
}

// DataReady checks the sensor to see if new data is available.
func (d *Device) DataReady() (bool, error) {
	if err := d.sendCommandWithResult(CmdDataReady, d.rx[0:3]); err != nil {
		return false, err
	}
	return !(d.rx[0]&0x07 == 0 && d.rx[1] == 0), nil
}

// StartPeriodicMeasurement puts the sensor into working mode, about 5s per measurement.
func (d *Device) StartPeriodicMeasurement() error {
	return d.sendCommand(CmdStartPeriodicMeasurement)
}

// StopPeriodicMeasurement stops the sensor reading data.
func (d *Device) StopPeriodicMeasurement() error {
	return d.sendCommand(CmdStopPeriodicMeasurement)
}

// StartLowPowerPeriodicMeasurement puts the sensor into low power working mode,
// about 30s per measurement.
func (d *Device) StartLowPowerPeriodicMeasurement() error {
	return d.sendCommand(CmdStartLowPowerPeriodicMeasurement)
}

// ReadData reads the data from the sensor and caches it.
func (d *Device) ReadData() error {
	if err := d.sendCommandWithResult(CmdReadMeasurement, d.rx[0:9]); err != nil {
		return err
	}
	d.co2 = binary.BigEndian.Uint16(d.rx[0:2])
	d.temperature = binary.BigEndian.Uint16(d.rx[3:5])
	d.humidity = binary.BigEndian.Uint16(d.rx[6:8])
	return nil
}

// ReadCO2 returns the CO2 concentration in PPM (parts per million).
func (d *Device) ReadCO2() (co2 int32, err error) {
	ok, err := d.DataReady()
	if err != nil {
		return 0, err
	}
	if ok {
		err = d.ReadData()
	}
	return int32(d.co2), err
}

// ReadTemperature returns the temperature in celsius milli degrees (°C/1000)
func (d *Device) ReadTemperature() (temperature int32, err error) {
	ok, err := d.DataReady()
	if err != nil {
		return 0, err
	}
	if ok {
		err = d.ReadData()
	}
	// temp = -45 + 175 * value / 2¹⁶
	return (-1 * 45000) + (21875 * (int32(d.temperature)) / 8192), err
}

// ReadTempC returns the value in the temperature value in Celsius.
func (d *Device) ReadTempC() float32 {
	t, _ := d.ReadTemperature()
	return float32(t) / 1000
}

// ReadTempF returns the value in the temperature value in Fahrenheit.
func (d *Device) ReadTempF() float32 {
	return d.ReadTempC()*1.8 + 32.0
}

// ReadHumidity returns the current relative humidity in %rH.
func (d *Device) ReadHumidity() (humidity int32, err error) {
	ok, err := d.DataReady()
	if err != nil {
		return 0, err
	}
	if ok {
		err = d.ReadData()
	}
	// humidity = 100 * value / 2¹⁶
	return (25 * int32(d.humidity)) / 16384, err
}

func (d *Device) sendCommand(command uint16) error {
	binary.BigEndian.PutUint16(d.tx[0:], command)
	return d.bus.Tx(uint16(d.Address), d.tx[0:2], nil)
}

func (d *Device) sendCommandWithValue(command, value uint16) error {
	binary.BigEndian.PutUint16(d.tx[0:], command)
	binary.BigEndian.PutUint16(d.tx[2:], value)
	d.tx[4] = crc8(d.tx[2:4])
	return d.bus.Tx(uint16(d.Address), d.tx[0:5], nil)
}

func (d *Device) sendCommandWithResult(command uint16, result []byte) error {
	binary.BigEndian.PutUint16(d.tx[0:], command)
	if err := d.bus.Tx(uint16(d.Address), d.tx[0:2], nil); err != nil {
		return err
	}
	time.Sleep(time.Millisecond)
	return d.bus.Tx(uint16(d.Address), nil, result)
}

func crc8(buf []byte) uint8 {
	var crc uint8 = 0xff
	for _, b := range buf {
		crc ^= b
		for i := 0; i < 8; i++ {
			if crc&0x80 != 0 {
				crc = (crc << 1) ^ 0x31
			} else {
				crc <<= 1
			}
		}
	}
	return crc & 0xff
}
