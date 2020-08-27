// Package sht3x provides a driver for the SHT3x digital humidity sensor
// series by Sensirion.
//
// Datasheet:
// https://www.sensirion.com/fileadmin/user_upload/customers/sensirion/Dokumente/0_Datasheets/Humidity/Sensirion_Humidity_Sensors_SHT3x_Datasheet_digital.pdf
//
package sht3x // import "tinygo.org/x/drivers/sht3x"

import (
	"time"

	"tinygo.org/x/drivers"
)

// Device wraps an I2C connection to a SHT31 device.
type Device struct {
	bus     drivers.I2C
	Address uint16
}

// New creates a new SHT31 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not initialize the device.
// You must call Configure() first in order to use the device itself.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: AddressA,
	}
}

// Read returns the temperature in celsius milli degrees (°C/1000).
func (d *Device) ReadTemperature() (tempMilliCelsius int32, err error) {
	tempMilliCelsius, _, err = d.ReadTemperatureHumidity()
	return tempMilliCelsius, err
}

// Read returns the relative humidity in hundredths of a percent.
func (d *Device) ReadHumidity() (relativeHumidity int16, err error) {
	_, relativeHumidity, err = d.ReadTemperatureHumidity()
	return relativeHumidity, err
}

// Read returns both the temperature and relative humidity.
func (d *Device) ReadTemperatureHumidity() (tempMilliCelsius int32, relativeHumidity int16, err error) {
	var rawTemp, rawHum, errx = d.rawReadings()
	if errx != nil {
		err = errx
		return
	}
	tempMilliCelsius = (35000 * int32(rawTemp) / 13107) - 45000
	relativeHumidity = int16(2000 * int32(rawHum) / 13107)
	return tempMilliCelsius, relativeHumidity, err
}

// rawReadings returns the sensor's raw values of the temperature and humidity
func (d *Device) rawReadings() (uint16, uint16, error) {
	d.bus.Tx(d.Address, []byte{MEASUREMENT_COMMAND_MSB, MEASUREMENT_COMMAND_LSB}, nil)

	time.Sleep(17 * time.Millisecond)

	var data [5]byte
	d.bus.Tx(d.Address, []byte{}, data[:])
	// ignore crc for now

	return readUint(data[0], data[1]), readUint(data[3], data[4]), nil
}

// readUint converts two bytes to uint16
func readUint(msb byte, lsb byte) uint16 {
	return (uint16(msb) << 8) | uint16(lsb)
}
