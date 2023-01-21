// Package wire implements the Dallas Semiconductor Corp.'s 1-wire bus system.
//
// Wikipedia: https://en.wikipedia.org/wiki/1-Wire
package wire // import "tinygo.org/x/drivers/1-wire"

import (
	"errors"
	"machine"
	"time"
)

// 1-Wire ROM commands
const (
	ONEWIRE_SEARCH_ROM   uint8 = 0xF0
	ONEWIRE_READ_ROM     uint8 = 0x33
	ONEWIRE_MATCH_ROM    uint8 = 0x55
	ONEWIRE_SKIP_ROM     uint8 = 0xCC
	ONEWIRE_ALARM_SEARCH uint8 = 0xEC
)

// Device wraps the data needed for the 1-wire protocol
type Device struct {
	Pin machine.Pin
}

var (
	errNoPresence = errors.New("Error: 1-Wire. No devices on the bus.")
)

// New. Creates a new 1-Wire connection.
// The pin must be pulled up to the VCC via a resistor (default 4.7k)
func New(pin machine.Pin) Device {
	return Device{
		Pin: pin,
	}
}

// Configure initializes the protocol, left for compatibility reasons
func (d Device) Configure() {

}

// Reset pulls DQ line low, then up.
func (d Device) Reset() error {
	d.Pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	time.Sleep(480 * time.Microsecond)
	d.Pin.Configure(machine.PinConfig{Mode: machine.PinInput})
	time.Sleep(70 * time.Microsecond)
	precence := d.Pin.Get()
	time.Sleep(410 * time.Microsecond)
	if precence {
		return errNoPresence
	}
	return nil
}

// Write transmits a byte as bit array to 1-Wire bus.
func (d Device) Write(data uint8) {
	for i := 0; i < 8; i++ {
		d.Pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		if data&1 == 1 {
			time.Sleep(6 * time.Microsecond)
			d.Pin.Configure(machine.PinConfig{Mode: machine.PinInput})
			time.Sleep(64 * time.Microsecond)
		} else {
			time.Sleep(60 * time.Microsecond)
			d.Pin.Configure(machine.PinConfig{Mode: machine.PinInput})
			time.Sleep(10 * time.Microsecond)
		}
		data >>= 1
	}
}

// Read receives a byte from 1-Wire bus
func (d Device) Read() (data uint8) {
	for i := 0; i < 8; i++ {
		data >>= 1
		d.Pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		time.Sleep(3 * time.Microsecond)
		d.Pin.Configure(machine.PinConfig{Mode: machine.PinInput})
		time.Sleep(8 * time.Microsecond)
		if d.Pin.Get() {
			data |= (1 << 7)
		}
		time.Sleep(60 * time.Microsecond)
	}
	return data
}

// Crc8 computes a Dallas Semiconductor 8 bit CRC.
func Сrc8(buffer []uint8, size int) (crc uint8) {
	// Dow-CRC using polynomial X^8 + X^5 + X^4 + X^0
	// Tiny 2x16 entry CRC table created by Arjen Lentz
	// See http://lentz.com.au/blog/calculating-crc-with-a-tiny-32-entry-lookup-table
	crc8_table := [...]uint8{
		0x00, 0x5E, 0xBC, 0xE2, 0x61, 0x3F, 0xDD, 0x83,
		0xC2, 0x9C, 0x7E, 0x20, 0xA3, 0xFD, 0x1F, 0x41,
		0x00, 0x9D, 0x23, 0xBE, 0x46, 0xDB, 0x65, 0xF8,
		0x8C, 0x11, 0xAF, 0x32, 0xCA, 0x57, 0xE9, 0x74,
	}
	for i := 0; i < size; i++ {
		crc = buffer[i] ^ crc // just re-using crc as intermediate
		crc = crc8_table[crc&0x0f] ^ crc8_table[16+((crc>>4)&0x0f)]
	}
	return crc
}
