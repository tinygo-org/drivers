// Package wire implements the Dallas Semiconductor Corp.'s 1-wire bus system.
//
// Wikipedia: https://en.wikipedia.org/wiki/1-Wire
package onewire // import "tinygo.org/x/drivers/onewire"

import (
	"errors"
	"machine"
	"time"
)

// OneWire ROM commands
const (
	READ_ROM   uint8 = 0x33
	MATCH_ROM  uint8 = 0x55
	SKIP_ROM   uint8 = 0xCC
	SEARCH_ROM uint8 = 0xF0
)

// Device wraps a connection to an 1-Wire devices.
type Device struct {
	p machine.Pin
}

// Config wraps a configuration to an 1-Wire devices.
type Config struct{}

// Errors list
var (
	errNoPresence  = errors.New("Error: OneWire. No devices on the bus.")
	errReadAddress = errors.New("Error: OneWire. Read address error: CRC mismatch.")
)

// New creates a new GPIO 1-Wire connection.
// The pin must be pulled up to the VCC via a resistor greater than 500 ohms (default 4.7k).
func New(p machine.Pin) Device {
	return Device{
		p: p,
	}
}

// Configure initializes the protocol.
func (d *Device) Configure(config Config) {}

// Reset pull DQ line low, then up.
func (d Device) Reset() error {
	d.p.Configure(machine.PinConfig{Mode: machine.PinOutput})
	time.Sleep(480 * time.Microsecond)
	d.p.Configure(machine.PinConfig{Mode: machine.PinInput})
	time.Sleep(70 * time.Microsecond)
	precence := d.p.Get()
	time.Sleep(410 * time.Microsecond)
	if precence {
		return errNoPresence
	}
	return nil
}

// WriteBit transmits a bit to 1-Wire bus.
func (d Device) WriteBit(data uint8) {
	d.p.Configure(machine.PinConfig{Mode: machine.PinOutput})
	if data&1 == 1 { // Send '1'
		time.Sleep(5 * time.Microsecond)
		d.p.Configure(machine.PinConfig{Mode: machine.PinInput})
		time.Sleep(60 * time.Microsecond)
	} else { // Send '0'
		time.Sleep(60 * time.Microsecond)
		d.p.Configure(machine.PinConfig{Mode: machine.PinInput})
		time.Sleep(5 * time.Microsecond)
	}
}

// Write transmits a byte as bit array to 1-Wire bus. (LSB first)
func (d Device) Write(data uint8) {
	for i := 0; i < 8; i++ {
		d.WriteBit(data)
		data >>= 1
	}
}

// ReadBit receives a bit from 1-Wire bus.
func (d Device) ReadBit() (data uint8) {
	d.p.Configure(machine.PinConfig{Mode: machine.PinOutput})
	time.Sleep(3 * time.Microsecond)
	d.p.Configure(machine.PinConfig{Mode: machine.PinInput})
	time.Sleep(8 * time.Microsecond)
	if d.p.Get() {
		data = 1
	}
	time.Sleep(60 * time.Microsecond)
	return data
}

// Read receives a byte from 1-Wire bus. (LSB first)
func (d Device) Read() (data uint8) {
	for i := 0; i < 8; i++ {
		data >>= 1
		data |= d.ReadBit() << 7
	}
	return data
}

// ReadAddress receives a 64-bit unique ROM ID from Device. (LSB first)
// Note: use this if there is only one slave device on the bus.
func (d Device) ReadAddress() ([]uint8, error) {
	var romid = make([]uint8, 8)
	if err := d.Reset(); err != nil {
		return nil, err
	}
	d.Write(READ_ROM)
	for i := 0; i < 8; i++ {
		romid[i] = d.Read()
	}
	if d.Сrc8(romid, 7) != romid[7] {
		return nil, errReadAddress
	}
	return romid, nil
}

// Select selects the address of the device for communication
func (d Device) Select(romid []uint8) error {
	if err := d.Reset(); err != nil {
		return err
	}
	if len(romid) == 0 {
		d.Write(SKIP_ROM)
		return nil
	}
	d.Write(MATCH_ROM)
	for i := 0; i < 8; i++ {
		d.Write(romid[i])
	}
	return nil
}

// Search searches for all devices on the bus.
// Note: max 32 slave devices per bus
func (d Device) Search(cmd uint8) ([][]uint8, error) {
	var (
		bit, bit_c  uint8 = 0, 0
		bitOffset   uint8 = 0
		lastZero    uint8 = 0
		lastFork    uint8 = 0
		lastAddress       = make([]uint8, 8)
		romIDs            = make([][]uint8, 32) //
		romIndex    uint8 = 0
	)

	for i := range romIDs {
		romIDs[i] = make([]uint8, 8)
	}

	for ok := true; ok; ok = (lastFork != 0) {
		if err := d.Reset(); err != nil {
			return nil, err
		}

		// send search command to bus
		d.Write(cmd)

		lastZero = 0

		for bitOffset = 0; bitOffset < 64; bitOffset++ {
			bit = d.ReadBit()   // read first address bit
			bit_c = d.ReadBit() // read second (complementary) address bit

			if bit == 1 && bit_c == 1 { // no device
				return nil, errNoPresence
			}

			if bit == 0 && bit_c == 0 { // collision
				if bitOffset == lastFork {
					bit = 1
				}
				if bitOffset < lastFork {
					bit = (lastAddress[bitOffset>>3] >> (bitOffset & 0x07)) & 1
				}
				if bit == 0 {
					lastZero = bitOffset
				}
			}

			if bit == 0 {
				lastAddress[bitOffset>>3] &= ^(1 << (bitOffset & 0x07))
			} else {
				lastAddress[bitOffset>>3] |= (1 << (bitOffset & 0x07))
			}
			d.WriteBit(bit)
		}
		lastFork = lastZero
		copy(romIDs[romIndex], lastAddress)
		romIndex++
	}
	return romIDs[:romIndex:romIndex], nil
}

// Crc8 compute a Dallas Semiconductor 8 bit CRC.
func (d Device) Сrc8(buffer []uint8, size int) (crc uint8) {
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
