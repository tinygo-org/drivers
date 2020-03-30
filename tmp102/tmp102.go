// Package tmp102 implements a driver for the TMP102 digital temperature sensor.
//
// Datasheet: https://download.mikroe.com/documents/datasheets/tmp102-data-sheet.pdf

package tmp102 // import "tinygo.org/x/drivers/tmp102"

import (
	"machine"
)

type TemperatureUnit int

const (
	UNIT_CELSIUS TemperatureUnit = iota
	UNIT_FARENHEIT
)

// Device holds the already configured I2C bus, the I2C
// address of the TMP102 and the temperature unit.
type Device struct {
	bus     machine.I2C
	address uint8
	unit    TemperatureUnit
}

// Config is the configuration for the TMP102.
type Config struct {
	Address uint8
	Unit    TemperatureUnit
}

// New creates a new TMP102 connection. The I2C bus must already be configured.
func New(bus machine.I2C) Device {
	return Device{
		bus: bus,
	}
}

// Configure initializes the sensor with the given parameters.
func (d *Device) Configure(cfg Config) {
	if cfg.Unit == 0 {
		cfg.Unit = UNIT_CELSIUS
	}

	if cfg.Address == 0 {
		cfg.Address = 0x48
	}

	d.address = cfg.Address
	d.unit = cfg.Unit
}

// Reads the temperature from the sensor and returns it in the configured temperature unit.
func (d *Device) ReadTemperature() float32 {

	tmpData := make([]byte, 2)

	d.bus.ReadRegister(d.address, 0x0, tmpData)

	temperatureSum := int((int16(tmpData[0])<<8 | int16(tmpData[1])) >> 4)

	if (temperatureSum & int(1<<11)) == int(1<<11) {
		temperatureSum |= int(0xf800)
	}

	temperature := float32(temperatureSum) * 0.0625

	if d.unit == UNIT_FARENHEIT {
		temperature = temperature*9.0/5.0 + 32
	}

	return temperature
}
