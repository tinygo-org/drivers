// Package max7219 implements an interface to a Maxim Integrated MAX7219 display-driver chip driving an 8x8 LED Matrix
//
// Datasheet: https://www.maximintegrated.com/en/products/power/display-power-control/MAX7219.html
//
package max7219

import (
	"github.com/tinygo-org/tinygo/src/machine"
)

// Uses a 3-wire serial interface
type Device struct {
	Data     machine.Pin // DIN
	Load     machine.Pin // Can also be labeled CS
	Clock    machine.Pin // CLK
	MaxInUse int         // Number of daisy chained MAX units
}

// Initialize the pins for a MAX7219 device as output pins
func New(pd machine.Pin, pl machine.Pin, pc machine.Pin, n ...int) *Device {

	pl.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pd.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pc.Configure(machine.PinConfig{Mode: machine.PinOutput})

	inUse := 1
	if len(n) > 0 {
		inUse = n[0]
	}

	dev := Device{Data: pd, Load: pl, Clock: pc, MaxInUse: inUse}

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

func (d Device) MaxAll(row byte, col byte) {
	d.Load.Low() // begin
	for c := 1; c <= d.MaxInUse; c++ {
		d.putByte(row)
		d.putByte(col)
	}
	d.Load.Low()
	d.Load.High()
}

// Address different MAX chips while having a couple cascaded
func (d Device) MaxOne(index int, row byte, col byte) {
	d.Load.Low()

	for c := d.MaxInUse; c > index; c-- {
		d.putByte(0) // NOP
		d.putByte(0) // NOP
	}

	d.putByte(row)
	d.putByte(col)

	for c := index - 1; c >= 1; c-- {
		d.putByte(0) // NOP
		d.putByte(0) // NOP
	}
}
