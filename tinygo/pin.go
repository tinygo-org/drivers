//go:build baremetal && tinygo

package tinygo // import "tinygo.org/x/drivers/tinygo"

import "machine"

// tinygo.Pin wraps machine.Pin to ensure correct pin mode is set when Get or Set methods are called.

type Pin struct {
	pin  machine.Pin
	mode machine.PinMode

	modeSet   bool
	inputMode machine.PinMode
}

func New(pin machine.Pin) *Pin {
	return &Pin{pin: pin, inputMode: machine.PinInput}
}

func NewWithInputMode(pin machine.Pin, inputMode machine.PinMode) *Pin {
	return &Pin{pin: pin, inputMode: inputMode}
}

func (p *Pin) Get() bool {
	if !p.modeSet || p.mode != p.inputMode {
		p.pin.Configure(machine.PinConfig{Mode: p.inputMode})
		p.mode = machine.PinInput
	}
	return p.pin.Get()
}

func (p *Pin) Set(high bool) {
	if !p.modeSet || p.mode != machine.PinOutput {
		p.pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		p.mode = machine.PinOutput
	}
	p.pin.Set(high)
}
