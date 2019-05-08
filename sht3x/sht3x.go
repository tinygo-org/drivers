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
		Address: Address,
	}
}

// Read returns the temperature in celsius and relative humidity.
func (d *Device) Read() (tempCelsius float32, relativeHumidity float32) {
	rawTemp, rawHum := d.rawReadings()
	tempCelsius = -45 + (175 * rawTemp / 65535)
  relativeHumidity = 100 * rawHum / 65535
	return tempCelsius, relativeHumidity
}

// rawReadings returns the sensor's raw values of the temperature and humidity
func (d *Device) rawReadings() (uint16, uint16) {
	d.bus.start(uint8(d.Address, false)
	d.bus.writeByte(MEASUREMENT_COMMAND_MSB)
	d.bus.writeByte(MEASUREMENT_COMMAND_LSB)
	d.bus.stop()

	time.Sleep(16 * time.Millisecond)

	data := make([]byte, 4)
	d.bus.start(uint8(d.Address, true)
	temp_msb = d.bus.readByte()
	temp_lsb = d.bus.readByte()
	d.bus.readByte() // skip crc
	hum_msb = d.bus.readByte()
	hum_lsb = d.bus.readByte()
	// ignore crc
	d.bus.stop()

	return readUint(temp_msb, temp_lsb), readUint(hum_msb, hum_lsb)
}

// readUint converts two bytes to uint16
func readUint(msb byte, lsb byte) uint16 {
	println("msb: ", msb)
	println("lsb: ", lsb)
	return (uint16(msb) << 8) | uint16(lsb)
}
