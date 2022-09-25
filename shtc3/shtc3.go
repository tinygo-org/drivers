// Package shtc3 provides a driver for the SHTC3 digital humidity sensor
// series by Sensirion.
//
// Datasheet:
// https://www.sensirion.com/fileadmin/user_upload/customers/sensirion/Dokumente/2_Humidity_Sensors/Datasheets/Sensirion_Humidity_Sensors_SHTC3_Datasheet.pdf
package shtc3 // import "tinygo.org/x/drivers/shtc3"

import (
	"time"

	"tinygo.org/x/drivers"
)

// Device wraps an I2C connection to a SHT31 device.
type Device struct {
	bus drivers.I2C
}

// New creates a new SHTC3 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not initialize the device.
// You must call Configure() first in order to use the device itself.
func New(bus drivers.I2C) Device {
	return Device{
		bus: bus,
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
	tempMilliCelsius = ((21875 * int32(rawTemp)) >> 13) - 45000
	relativeHumidity = int16((1250 * int32(rawHum)) >> 13)
	return tempMilliCelsius, relativeHumidity, err
}

// rawReadings returns the sensor's raw values of the temperature and humidity
func (d *Device) rawReadings() (uint16, uint16, error) {
	var data [6]byte
	d.bus.Tx(SHTC3_ADDRESS, []byte(SHTC3_CMD_MEASURE_HP), data[:])
	// ignore crc for now
	return readUint(data[0], data[1]), readUint(data[3], data[4]), nil
}

// WakeUp makes device leave sleep mode
func (d *Device) WakeUp() error {
	d.bus.Tx(SHTC3_ADDRESS, []byte(SHTC3_CMD_WAKEUP), nil)
	time.Sleep(1 * time.Millisecond)
	return nil
}

// Sleep makes device go to sleep
func (d *Device) Sleep() error {
	d.bus.Tx(SHTC3_ADDRESS, []byte(SHTC3_CMD_SLEEP), nil)
	return nil
}

// readUint converts two bytes to uint16
func readUint(msb byte, lsb byte) uint16 {
	return (uint16(msb) << 8) | uint16(lsb)
}
