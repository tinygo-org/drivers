// Package ds1307 provides a driver for the DS1307 RTC
//
// Datasheet:
// https://datasheets.maximintegrated.com/en/ds/DS1307.pdf
package ds1307 // import "tinygo.org/x/drivers/ds1307"

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// Device wraps an I2C connection to a DS1307 device.
type Device struct {
	bus         drivers.I2C
	Address     uint8
	AddressSRAM uint8
}

// New creates a new DS1307 connection. I2C bus must be already configured.
func New(bus drivers.I2C) Device {
	return Device{bus: bus,
		Address:     uint8(I2CAddress),
		AddressSRAM: SRAMBeginAddres,
	}
}

// SetTime sets the time and date
func (d *Device) SetTime(t time.Time) error {
	data := make([]byte, 8)
	data[0] = uint8(TimeDate)
	data[1] = decToBcd(t.Second())
	data[2] = decToBcd(t.Minute())
	data[3] = decToBcd(t.Hour())
	data[4] = decToBcd(int(t.Weekday() + 1))
	data[5] = decToBcd(t.Day())
	data[6] = decToBcd(int(t.Month()))
	data[7] = decToBcd(t.Year() - 2000)
	err := d.bus.Tx(uint16(d.Address), data, nil)
	return err
}

// ReadTime returns the date and time
func (d *Device) ReadTime() (time.Time, error) {
	data := make([]byte, 8)
	err := legacy.ReadRegister(d.bus, d.Address, uint8(TimeDate), data)
	if err != nil {
		return time.Time{}, err
	}
	seconds := bcdToDec(data[0] & 0x7F)
	minute := bcdToDec(data[1])
	hour := hoursBCDToInt(data[2])
	day := bcdToDec(data[4])
	month := time.Month(bcdToDec(data[5]))
	year := bcdToDec(data[6])
	year += 2000

	t := time.Date(year, month, day, hour, minute, seconds, 0, time.UTC)
	return t, nil
}

// Seek sets the offset for the next Read or Write on SRAM to offset, interpreted
// according to whence: 0 means relative to the origin of the SRAM, 1 means
// relative to the current offset, and 2 means relative to the end.
// returns new offset and error, if any
func (d *Device) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		whence = SRAMBeginAddres
	case 1:
		whence = int(d.AddressSRAM)
	case 2:
		whence = SRAMEndAddress
	default:
		return 0, errors.New("invalid starting point")
	}
	d.AddressSRAM = uint8(whence) + uint8(offset)
	if d.AddressSRAM > SRAMEndAddress {
		return 0, errors.New("EOF")
	}
	return int64(d.AddressSRAM), nil
}

// Write writes len(data) bytes to SRAM
// returns number of bytes written and error, if any
func (d *Device) Write(data []byte) (n int, err error) {
	if int(d.AddressSRAM)+len(data)-1 > SRAMEndAddress {
		return 0, errors.New("writing outside of SRAM")
	}
	buffer := make([]byte, len(data)+1)
	buffer[0] = d.AddressSRAM
	copy(buffer[1:], data)
	err = d.bus.Tx(uint16(d.Address), buffer, nil)
	if err != nil {
		return 0, err
	}
	d.Seek(int64(len(data)), 1)
	return len(data), nil
}

// Read reads len(data) from SRAM
// returns number of bytes written and error, if any
func (d *Device) Read(data []uint8) (n int, err error) {
	if int(d.AddressSRAM)+len(data)-1 > SRAMEndAddress {
		return 0, errors.New("EOF")
	}
	err = legacy.ReadRegister(d.bus, d.Address, d.AddressSRAM, data)
	if err != nil {
		return 0, err
	}
	d.Seek(int64(len(data)), 1)
	return len(data), nil
}

// SetOscillatorFrequency sets output oscillator frequency
// Available modes: SQW_OFF, SQW_1HZ, SQW_4KHZ, SQW_8KHZ, SQW_32KHZ
func (d *Device) SetOscillatorFrequency(sqw uint8) error {
	data := []byte{uint8(Control), sqw}
	err := d.bus.Tx(uint16(d.Address), data, nil)
	return err
}

// IsOscillatorRunning returns if the oscillator is running
func (d *Device) IsOscillatorRunning() bool {
	data := []byte{0}
	err := legacy.ReadRegister(d.bus, d.Address, uint8(TimeDate), data)
	if err != nil {
		return false
	}
	return (data[0] & (1 << CH)) == 0
}

// SetOscillatorRunning starts/stops internal oscillator by toggling halt bit
func (d *Device) SetOscillatorRunning(running bool) error {
	data := make([]byte, 3)
	err := legacy.ReadRegister(d.bus, d.Address, uint8(TimeDate), data)
	if err != nil {
		return err
	}
	if running {
		data[0] &^= (1 << CH)
	} else {
		data[0] |= (1 << CH)
	}
	data[1], data[0] = data[0], uint8(TimeDate)
	err = d.bus.Tx(uint16(d.Address), data[:2], nil)
	return err
}

// decToBcd converts int to BCD
func decToBcd(dec int) uint8 {
	return uint8(dec + 6*(dec/10))
}

// bcdToDec converts BCD to int
func bcdToDec(bcd uint8) int {
	return int(bcd - 6*(bcd>>4))
}

// hoursBCDToInt converts the BCD hours to int
func hoursBCDToInt(value uint8) (hour int) {
	if value&0x40 != 0x00 {
		hour = bcdToDec(value & 0x1F)
		if (value & 0x20) != 0x00 {
			hour += 12
		}
	} else {
		hour = bcdToDec(value)
	}
	return
}
