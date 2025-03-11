// Datasheet: https://www.analog.com/media/en/technical-documentation/data-sheets/max6675.pdf
package max6675

import (
	"errors"
	"machine"

	"tinygo.org/x/drivers"
)

// ErrThermocoupleOpen is returned when the thermocouple input is open.
// i.e. not attached  or faulty
var ErrThermocoupleOpen = errors.New("thermocouple input open")

type Device struct {
	bus drivers.SPI
	cs  machine.Pin
}

// Create a new Device to read from a MAX6675 thermocouple.
// Pins must be configured before use.  Frequency for SPI
// should be 4.3MHz maximum.
func NewDevice(bus drivers.SPI, cs machine.Pin) *Device {
	return &Device{
		bus: bus,
		cs:  cs,
	}
}

// Read and return the temperature in celsius
func (d *Device) Read() (float32, error) {
	var (
		read  []byte = []byte{0, 0}
		value uint16
	)

	d.cs.Low()
	if err := d.bus.Tx([]byte{0, 0}, read); err != nil {
		return 0, err
	}
	d.cs.High()

	// datasheet: Bit D2 is normally low and goes high if the thermocouple input is open.
	if read[1]&0x04 == 0x04 {
		return 0, ErrThermocoupleOpen
	}

	// data is 12 bits, split across the two bytes
	// -XXXXXXX XXXXX---
	value = (uint16(read[0]) << 5) | (uint16(read[1]) >> 3)

	return float32(value) * 0.25, nil
}
