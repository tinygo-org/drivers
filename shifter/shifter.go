// Package shifter is for 8bit shift register
package shifter // import "tinygo.org/x/drivers/shifter"

import (
	"machine"
)

// Device holds the pins.
type Device struct {
	latch machine.Pin
	clk   machine.Pin
	out   machine.Pin
}

// New returns a new thermistor driver given an ADC pin.
func New(latch, clk, out machine.Pin) Device {
	return Device{
		latch: latch,
		clk:   clk,
		out:   out,
	}
}

// Configure here just for interface compatibility.
func (d *Device) Configure() {
	d.latch.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.clk.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.out.Configure(machine.PinConfig{Mode: machine.PinInput})
}

// Read8Input reads the 8 inputs and return an uint8
func (d *Device) Read8Input() uint8 {
	return uint8(d.readInput(8))
}

// Read16Input reads the 16 inputs and return an uint16
func (d *Device) Read16Input() uint16 {
	return uint16(d.readInput(16))
}

// Read32Input reads the 32 inputs and return an uint32
func (d *Device) Read32Input() uint32 {
	return d.readInput(32)
}

// readInput reads howMany bits from the shift register
func (d *Device) readInput(howMany int8) uint32 {
	d.latch.High()
	var data uint32
	for i := howMany - 1; i >= 0; i-- {
		d.clk.Low()
		if d.out.Get() {
			data |= 1 << i
		}
		d.clk.High()
	}
	d.latch.Low()
	return data
}
