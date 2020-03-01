// Package shiftregister is for 8bit shift output register using 3 GPIO pins
//
// Tested with SN74HC595 with 8 outputs
// Wiring:
// 	 Latch connected to ST_CP of 74HC595 (12)
//   Clock connected to SH_CP of 74HC595 (11)
//   Out:  connected to DS of 74HC595 (14)
// Datsheet https://www.ti.com/lit/ds/symlink/sn74hc595.pdf
package shiftregister

import (
	"machine"
)

type NumberBit int8

// Bit number of the register
const (
	EIGHT_BITS     NumberBit = 8
	SIXTEEN_BITS   NumberBit = 16
	THIRTYTWO_BITS NumberBit = 32
)

// Device holds pin number
type Device struct {
	latch, clock, out machine.Pin
	bits              NumberBit
}

// DeviceConfiguration for setting up the shift register
type DeviceConfiguration struct {
	Bits              NumberBit
	Latch, Clock, Out machine.Pin
}

// New returns a new shift output register device
func New(Bits NumberBit, Latch, Clock, Out machine.Pin) *Device {
	return &Device{
		latch: Latch,
		clock: Clock,
		out:   Out,
		bits:  Bits,
	}
}

// Configure set hardware configuration
func (d *Device) Configure() {
	d.latch.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.clock.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.out.Configure(machine.PinConfig{Mode: machine.PinInput})
	d.latch.High()
}

// WriteMask applies mask's bits to register's outputs pin
// mask's MSB set Q1, LSB set Q8 (for 8 bits mask)
func (d *Device) WriteMask(mask uint32) {
	d.out.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.latch.Low()
	for i := 0; i < int(d.bits); i++ {
		d.clock.Low()
		d.out.Set(mask&1 != 0)
		mask = mask >> 1
		d.clock.High()
	}
	d.latch.High()
}
