package ds18b20

import (
	"errors"
	"tinygo.org/x/drivers/onewire"
)

// Device ROM commands
const (
	DS18B20_CONVERT_TEMPERATURE uint8 = 0x44
	DS18B20_READ_SCRATCHPAD     uint8 = 0xBE
	DS18B20_COPY_SCRATCHPAD     uint8 = 0x48
	DS18B20_WRITE_SCRATCHPAD    uint8 = 0x4E
	DS18B20_READ_POWER_SUPPLY   uint8 = 0xB4
	DS18B20_RECALL_E2           uint8 = 0xB8
)

// Device wraps the 1-wire protocol to a ds18b20 device.
type Device struct {
	owd        onewire.Device
	RomID      []uint8
	ScratchPad []uint8
}

// Errors list
var (
	errReadTemperature = errors.New("Error: DS18B20. Read temperature error: CRC mismatch.")
	errReadAddress     = errors.New("Error: DS18B20. Read address error: CRC mismatch.")
)

// New returns a ds18b20 device
func New(owd onewire.Device) Device {
	return Device{
		owd:        owd,
		RomID:      make([]uint8, 8),
		ScratchPad: make([]uint8, 9),
	}
}

// ThermometerResolution set thermometer resolution from 9 to 12 bits
func (d Device) ThermometerResolution(id, resolution uint8) error {
	if 9 <= resolution && resolution <= 12 {
		resolution = ((resolution - 9) << 5) | 0x1F
		err := d.owd.Reset()
		if err != nil {
			return err
		}
		d.addressRoutine()
		d.owd.Write(DS18B20_WRITE_SCRATCHPAD)
		d.owd.Write(0xFF)
		d.owd.Write(0x00)
		d.owd.Write(resolution)
	}
	return nil
}

// RequestTemperature sends request
func (d Device) RequestTemperature() error {
	if err := d.owd.Reset(); err != nil {
		return err
	}
	d.addressRoutine()
	d.owd.Write(DS18B20_CONVERT_TEMPERATURE)
	return nil
}

// ReadTemperature returns the temperature in celsius milli degrees (°C/1000)
//
// 1 bit degrees = 1/16 = 0.0625
//
// TEMPERATURE/DATA RELATIONSHIP
// +125.0000°C 0000 0111 1101 0000 07D0h
// +085.0000°C 0000 0101 0101 0000 0550h
// +025.0625°C 0000 0001 1001 0001 0191h
// +010.1250°C 0000 0000 1010 0010 00A2h
// +000.5000°C 0000 0000 0000 1000 0008h
// ±000.0000°C 0000 0000 0000 0000 0000h
// -000.5000°C 1111 1111 1111 1000 FFF8h
// -010.1250°C 1111 1111 0101 1110 FF5Eh
// -025.0625°C 1111 1110 0110 1111 FE6Fh
// -055.0000°C 1111 1100 1001 0000 FC90h
func (d Device) ReadTemperature() (temperature int32, err error) {
	if err := d.owd.Reset(); err != nil {
		return temperature, err
	}
	d.addressRoutine()
	d.owd.Write(DS18B20_READ_SCRATCHPAD)
	for i := 0; i < 9; i++ {
		d.ScratchPad[i] = d.owd.Read()
	}
	if onewire.Сrc8(d.ScratchPad, 8) != d.ScratchPad[8] {
		return temperature, errReadTemperature
	}
	temperature = int32(uint16(d.ScratchPad[0]) | uint16(d.ScratchPad[1])<<8)
	if temperature&0x8000 == 0x8000 {
		temperature -= 0x10000
	}
	return (temperature * 625 / 10), nil
}

// ReadAddress
func (d Device) ReadAddress() error {
	if err := d.owd.Reset(); err != nil {
		return err
	}
	d.owd.Write(onewire.ONEWIRE_READ_ROM)
	for i := 0; i < 8; i++ {
		d.RomID[i] = d.owd.Read()
	}
	if onewire.Сrc8(d.RomID, 7) != d.RomID[7] {
		return errReadAddress
	}
	return nil
}

func (d Device) addressRoutine() {
	d.owd.Write(onewire.ONEWIRE_SKIP_ROM)
}
