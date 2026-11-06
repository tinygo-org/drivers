//go:build tinygo && (rp2040 || rp2350 || stm32 || k210 || esp32 || esp32c3 || esp32s3 || esp32c6 || nrf || sam || (avr && (atmega328p || atmega328pb)))

// Implementation based on:
// https://gist.github.com/aykevl/3fc1683ed77bb0a9c07559dfe857304a

// Note: build constraints in this file list targets that define machine.PinToggle.
// If this is supported for additional targets in the future, they can be added above.

package encoders

import (
	"machine"
	"runtime/volatile"
)

var (
	states = []int8{0, -1, 1, 0, 1, 0, 0, -1, -1, 0, 0, 1, 0, 1, -1, 0}
)

type InterruptablePin[P any, CFT func(P)] interface {
	SetInterrupt(change machine.PinChange, callback CFT) error
	Configure(machine.PinConfig)
	Get() bool
}

// NewQuadratureViaInterrupt returns a rotary encoder device that uses GPIO
// interrupts and a lookup table to keep track of quadrature state changes.
//
// This constructur is only available for TinyGo targets for which machine.PinToggle
// is defined as a valid interrupt type.
func NewQuadratureViaInterrupt[P InterruptablePin[P, func(P)]](pinA, pinB P) *QuadratureDevice {
	return &QuadratureDevice{impl: &quadInterruptImpl[P]{pinA: pinA, pinB: pinB, oldAB: 0b00000011}}
}

type quadInterruptImpl[P InterruptablePin[P, func(P)]] struct {
	pinA P
	pinB P

	// precision int

	oldAB int
	value volatile.Register32
}

func (enc *quadInterruptImpl[P]) configure(cfg QuadratureConfig) error {
	enc.pinA.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	enc.pinA.SetInterrupt(machine.PinToggle, enc.interrupt)

	enc.pinB.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	enc.pinB.SetInterrupt(machine.PinToggle, enc.interrupt)

	return nil
}

func (enc *quadInterruptImpl[P]) interrupt(pin P) {
	aHigh, bHigh := enc.pinA.Get(), enc.pinB.Get()
	enc.oldAB <<= 2
	if aHigh {
		enc.oldAB |= 1 << 1
	}
	if bHigh {
		enc.oldAB |= 1
	}
	enc.writeValue(enc.readValue() + int(states[enc.oldAB&0x0f]))
}

// readValue gets the value using volatile operations and returns it as an int
func (enc *quadInterruptImpl[P]) readValue() int {
	return int(enc.value.Get())
}

// writeValue set the value to the specified int using volatile operations
func (enc *quadInterruptImpl[P]) writeValue(v int) {
	enc.value.Set(uint32(v))
}
