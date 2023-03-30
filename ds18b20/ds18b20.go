// Package ds18b20 provides a driver for the DS18B20 digital thermometer
//
// Datasheet:
// https://www.analog.com/media/en/technical-documentation/data-sheets/DS18B20.pdf
package ds18b20 // import "tinygo.org/x/drivers/ds18b20"

import (
	"errors"
)

// Device ROM commands
const (
	CONVERT_TEMPERATURE uint8 = 0x44
	READ_SCRATCHPAD     uint8 = 0xBE
	WRITE_SCRATCHPAD    uint8 = 0x4E
)

type OneWireDevice interface {
	Write(uint8)
	Read() uint8
	Select([]uint8) error
	Сrc8([]uint8, int) uint8
}

// Device wraps a connection to an 1-Wire devices.
type Device struct {
	owd OneWireDevice
}

// Errors list
var (
	errReadTemperature = errors.New("Error: DS18B20. Read temperature error: CRC mismatch.")
)

func New(owd OneWireDevice) Device {
	return Device{
		owd: owd,
	}
}

// Configure. Initializes the device, left for compatibility reasons.
func (d Device) Configure() {}

// ThermometerResolution sets thermometer resolution from 9 to 12 bits
func (d Device) ThermometerResolution(romid []uint8, resolution uint8) {
	if 9 <= resolution && resolution <= 12 {
		d.owd.Select(romid)
		d.owd.Write(WRITE_SCRATCHPAD)               // send three data bytes to scratchpad (TH, TL, and config)
		d.owd.Write(0xFF)                           // to TH
		d.owd.Write(0x00)                           // to TL
		d.owd.Write(((resolution - 9) << 5) | 0x1F) // to resolution config
	}
}

// RequestTemperature sends request to device
func (d Device) RequestTemperature(romid []uint8) {
	d.owd.Select(romid)
	d.owd.Write(CONVERT_TEMPERATURE)
}

// ReadTemperatureRaw returns the raw temperature.
// ScratchPad memory map:
// byte 0: Temperature LSB
// byte 1: Temperature MSB
func (d Device) ReadTemperatureRaw(romid []uint8) ([]uint8, error) {
	spb := make([]uint8, 9) // ScratchPad buffer
	d.owd.Select(romid)
	d.owd.Write(READ_SCRATCHPAD)
	for i := 0; i < 9; i++ {
		spb[i] = d.owd.Read()
	}
	if d.owd.Сrc8(spb, 8) != spb[8] {
		return nil, errReadTemperature
	}
	return spb[:2:2], nil
}

// ReadTemperature returns the temperature in celsius milli degrees (°C/1000)
func (d Device) ReadTemperature(romid []uint8) (int32, error) {
	raw, err := d.ReadTemperatureRaw(romid)
	if err != nil {
		return 0, err
	}
	t := int32(uint16(raw[0]) | uint16(raw[1])<<8)
	if t&0x8000 == 0x8000 {
		t -= 0x10000
	}
	return (t * 625 / 10), nil
}
