// Package shiftregister is for 8bit shift output register using 3 GPIO pins like SN74ALS164A, SN74AHC594, SN74AHC595, ...
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
	latch, clock, out machine.Pin // IC wiring
	bits              NumberBit   // Pin number
	mask              uint32      // keep all pins state
}

// ShiftPin is the implementation of the ShiftPin interface.
// ShiftPin provide an interface like regular machine.Pin
type ShiftPin struct {
	mask uint32  // Bit representing the pin
	d    *Device // Reference to the register
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
	d.out.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.latch.High()
}

// WriteMask applies mask's bits to register's outputs pin
// mask's MSB set Q1, LSB set Q8 (for 8 bits mask)
func (d *Device) WriteMask(mask uint32) {
	d.mask = mask // Keep the mask for individual addressing
	d.latch.Low()
	for i := 0; i < int(d.bits); i++ {
		d.clock.Low()
		d.out.Set(mask&1 != 0)
		mask = mask >> 1
		d.clock.High()
	}
	d.latch.High()
}

// GetShiftPin return an individually addressable pin
func (d *Device) GetShiftPin(pin int) *ShiftPin {
	if pin < 0 || pin > int(d.bits) {
		panic("invalid pin number")
	}
	return &ShiftPin{
		mask: 1 << pin,
		d:    d,
	}

}

// Set changes the value of this register pin.
func (p ShiftPin) Set(value bool) {
	d := p.d
	if value {
		d.WriteMask(d.mask | p.mask)
	} else {
		d.WriteMask(d.mask & ^p.mask)
	}
}

// High sets this shift register pin to high.
func (p ShiftPin) High() {
	p.Set(true)
}

// Low sets this shift register pin to low.
func (p ShiftPin) Low() {
	p.Set(false)
}
