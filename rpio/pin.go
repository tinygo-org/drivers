// Implementation of the Pin interface for the Raspberry Pi GPIO pins.
// Depends on the go-rpio library.
// Configures pin modes automatically when Get or Set methods are called.

package rpio // import "tinygo.org/x/drivers/rpio"

import go_rpio "github.com/stianeikeland/go-rpio/v4"

type Pin struct {
	modeSet bool
	Mode    go_rpio.Mode
	Pin     go_rpio.Pin
}

func NewPin(pinNumber int) *Pin {
	return &Pin{Pin: go_rpio.Pin(pinNumber)}
}

func (p *Pin) Get() bool {
	if !p.modeSet || p.Mode != go_rpio.Input {
		p.Pin.Input()
		p.Mode = go_rpio.Input
	}
	return p.Pin.Read() == go_rpio.High
}

func (p *Pin) Set(high bool) {
	if !p.modeSet || p.Mode != go_rpio.Output {
		p.Pin.Output()
		p.Mode = go_rpio.Output
	}
	state := go_rpio.Low
	if high {
		state = go_rpio.High
	}
	p.Pin.Write(state)
}
