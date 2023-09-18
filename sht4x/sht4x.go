// Package sht4x provides a driver for the SHT4x digital humidity sensor series by Sensirion.
// Datasheet: https://www.sensirion.com/media/documents/33FD6951/64D3B030/Sensirion_Datasheet_SHT4x.pdf
package sht4x

import (
	"time"

	"tinygo.org/x/drivers"
)

const DefaultAddress = 0x44

const (
	// single-shot, high-repeatability measurement
	commandMeasurement = 0xfd
)

// Device represents a SHT4x sensor
type Device struct {
	bus     drivers.I2C
	Address uint8
}

// New creates a new SHT4x connection. The I2C bus must already be
// configured.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: DefaultAddress,
	}
}

// ReadTemperatureHumidity starts a measurement and then reads out the results. This function blocks
// while the measurement is in progress.
//
// Temperature is returned in [degree Celsius], multiplied by 1000,
// and relative humidity in [percent relative humidity], multiplied by 1000.
func (d *Device) ReadTemperatureHumidity() (temperatureMilliCelsius int32, relativeHumidityMilliPercent int32, err error) {
	rawTemp, rawHum, err := d.rawReadings()
	if err != nil {
		return 0, 0, err
	}

	// from the reference driver: https://github.com/Sensirion/embedded-sht/blob/fcc8a523210cc1241a2750899ff6b0f68f3ed212/sht4x/sht4x.c#L81
	temperatureMilliCelsius = ((21875 * int32(rawTemp)) >> 13) - 45000
	relativeHumidityMilliPercent = ((15625 * int32(rawHum)) >> 13) - 6000

	return temperatureMilliCelsius, relativeHumidityMilliPercent, err
}

// rawReadings returns the sensor's raw values of the temperature and humidity
func (d *Device) rawReadings() (uint16, uint16, error) {
	err := d.bus.Tx(uint16(d.Address), []byte{commandMeasurement}, nil)
	if err != nil {
		return 0, 0, err
	}

	// max time for measurement according to datasheet
	time.Sleep(10 * time.Millisecond)

	var data [6]byte
	err = d.bus.Tx(uint16(d.Address), nil, data[:])
	if err != nil {
		return 0, 0, err
	}

	tTicks := readUint(data[0], data[1])
	rhTicks := readUint(data[3], data[4])

	return tTicks, rhTicks, nil
}

// readUint converts two bytes to uint16
func readUint(msb byte, lsb byte) uint16 {
	return (uint16(msb) << 8) | uint16(lsb)
}
