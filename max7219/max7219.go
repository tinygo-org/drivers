// Package max7219 implements an interface to a Maxim Integrated MAX7219 display-driver chip driving an 8x8 LED Matrix
//
// Datasheet: https://www.maximintegrated.com/en/products/power/display-power-control/MAX7219.html
//
package max7219

import (
	"machine"
)

// Uses a 3-wire serial interface
type Device struct {
	Data     machine.Pin // DIN
	Load     machine.Pin // Can also be labeled CS
	Clock    machine.Pin // CLK
	MaxInUse int         // How many MAX7219s are daisy-chained
}

// Initialize the pins for a MAX7219 device as output pins
func New(pd machine.Pin, pl machine.Pin, pc machine.Pin, n ...int) *Device {

	pl.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pd.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pc.Configure(machine.PinConfig{Mode: machine.PinOutput})

	numDevices := 1
	if len(n) > 0 {
		numDevices = n[0]
	}

	dev := Device{Data: pd, Load: pl, Clock: pc, MaxInUse: numDevices}

	return &dev
}

// Initialize the matrix for input
func (d Device) Configure() {
	d.MaxSingle(REG_SCANLIMIT, 0x07)
	d.MaxSingle(REG_DECODE_MODE, 0x00)
	d.MaxSingle(REG_SHUTDOWN, 0x01)
	d.MaxSingle(REG_DISPLAY_TEST, 0x00)
	// Wipe all rows of the matrix
	for r := 0; r <= 8; r++ {
		d.MaxSingle(byte(r), 0)
	}
	d.MaxSingle(REG_INTENSITY, 0x0F&0x0F)
}

// Helper function to send a single byte to the matrix
func (d Device) putByte(data byte) {
	for i := 0x08; i > 0; i-- {
		mask := byte(0x01 << (i - 1))
		d.Clock.Low() // tick
		if (data & mask) > 0 {
			d.Data.High()
		} else {
			d.Data.Low()
		}
		d.Clock.High() // tock
	}
}

// Write the bitstring `col` to `row`
// Note: Row is indexed at 1
func (d Device) MaxSingle(row byte, col byte) {
	d.Load.Low()
	d.putByte(row)
	d.putByte(col)
	d.Load.Low()
	d.Load.High()
}

// Write the bitstring to all cascaded-MAX7219 devices
func (d Device) MaxAll(row byte, col byte) {
	d.Load.Low() // start operation

	for i := 1; i <= d.MaxInUse; i++ {
		d.putByte(row)
		d.putByte(col)
	}

	// finish operation
	d.Load.Low()
	d.Load.High()
}

// Send data to a single MAX in a cascaded setup
func (d Device) MaxOne(m int, row byte, col byte) {
	// m: device index
	// row: row to address
	// col: bytestring to write to row

	d.Load.Low() // begin

	// Send 2x NOP to each MAX7219
	for c := d.MaxInUse; c > m; c-- {
		d.putByte(0)
		d.putByte(0)
	}

	// Send data
	d.putByte(row)
	d.putByte(col)

	// 2x NOP again
	for c := m - 1; c >= 1; c-- {
		d.putByte(0)
		d.putByte(0)
	}

	// Finish op
	d.Load.Low()
	d.Load.High()
}
