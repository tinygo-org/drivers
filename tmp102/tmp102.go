// Package tmp102 implements a driver for the TMP102 digital temperature sensor.
//
// Datasheet: https://download.mikroe.com/documents/datasheets/tmp102-data-sheet.pdf

package tmp102 // import "tinygo.org/x/drivers/tmp102"

import (
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// Device holds the already configured I2C bus and the address of the sensor.
type Device struct {
	bus     drivers.I2C
	address uint8
}

// Config is the configuration for the TMP102.
type Config struct {
	Address uint8
}

// New creates a new TMP102 connection. The I2C bus must already be configured.
func New(bus drivers.I2C) Device {
	return Device{
		bus: bus,
	}
}

// Configure initializes the sensor with the given parameters.
func (d *Device) Configure(cfg Config) {
	if cfg.Address == 0 {
		cfg.Address = Address
	}

	d.address = cfg.Address
}

// Connected checks if the config register can be read and that the configuration is correct.
func (d *Device) Connected() bool {
	configData := make([]byte, 2)
	err := legacy.ReadRegister(d.bus, d.address, RegConfiguration, configData)
	// Check the reset configuration values.
	if err != nil || configData[0] != 0x60 || configData[1] != 0xA0 {
		return false
	}
	return true

}

// Reads the temperature from the sensor and returns it in celsius milli degrees (°C/1000).
func (d *Device) ReadTemperature() (temperature int32, err error) {

	tmpData := make([]byte, 2)

	err = legacy.ReadRegister(d.bus, d.address, RegTemperature, tmpData)

	if err != nil {
		return
	}

	temperatureSum := int32((int16(tmpData[0])<<8 | int16(tmpData[1])) >> 4)

	if (temperatureSum & int32(1<<11)) == int32(1<<11) {
		temperatureSum |= int32(0xf800)
	}

	temperature = temperatureSum * 625

	return temperature / 10, nil
}
