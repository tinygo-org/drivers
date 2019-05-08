// Package sht3x provides a driver for the SHT3x digital humidity sensor
// series by Sensirion.
//
// Datasheet:
// https://www.sensirion.com/fileadmin/user_upload/customers/sensirion/Dokumente/0_Datasheets/Humidity/Sensirion_Humidity_Sensors_SHT3x_Datasheet_digital.pdf
//
package sht3x

import (
	"machine"
	"time"
)

// Device wraps an I2C connection to a SHT31 device.
type Device struct {
	bus     machine.I2C
	Address uint16
}

// New creates a new SHT31 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not initialize the device.
// You must call Configure() first in order to use the device itself.
func New(bus machine.I2C) Device {
	return Device{
		bus:     bus,
		Address: AddressA,
	}
}

// Read returns the temperature in celsius.
func (d *Device) ReadTemperature() (tempCelsius float32) {
	tempCelsius, _ = d.Read()
	return tempCelsius
}

// Read returns the relative humidity.
func (d *Device) ReadHumidity() (relativeHumidity float32) {
	_, relativeHumidity = d.Read()
	return relativeHumidity
}

// Read returns both the temperature in celsius and relative humidity.
func (d *Device) Read() (tempCelsius float32, relativeHumidity float32) {
	var rawTemp, rawHum = d.rawReadings()
	tempCelsius = -45.0 + (175.0 * float32(rawTemp) / 65535.0)
	relativeHumidity = 100.0 * float32(rawHum) / 65535.0
	return tempCelsius, relativeHumidity
}

// rawReadings returns the sensor's raw values of the temperature and humidity
func (d *Device) rawReadings() (uint16, uint16) {
	d.bus.Tx(d.Address, []byte{MEASUREMENT_COMMAND_MSB, MEASUREMENT_COMMAND_LSB}, nil)

	time.Sleep(17 * time.Millisecond)

	data := make([]byte, 5)
	d.bus.Tx(d.Address, []byte{}, data)
	// ignore crc for now

	return readUint(data[0], data[1]), readUint(data[3], data[4])
}

// readUint converts two bytes to uint16
func readUint(msb byte, lsb byte) uint16 {
	return (uint16(msb) << 8) | uint16(lsb)
}
