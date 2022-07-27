// Package makeybutton providers a driver for a button that can be triggered
// by anything that is conductive by using an ultra high value resistor.
//
// Inspired by the amazing MakeyMakey
// https://makeymakey.com/
//
package makeybutton

import (
	"machine"
	"time"
)

// ButtonState represents the state of a MakeyButton.
type ButtonState int

const (
	NeverPressed ButtonState = 0
	Press                    = 1
	Release                  = 2
)

// ButtonEvent represents when the state of a Button changes.
type ButtonEvent int

const (
	NotChanged ButtonEvent = 0
	Pressed                = 1
	Released               = 2
)

// Button is a "button"-like device that acts like a MakeyMakey.
type Button struct {
	pin              machine.Pin
	state            ButtonState
	readings         *Buffer
	HighMeansPressed bool
}

// NewButton creates a new Button.
func NewButton(pin machine.Pin) *Button {
	return &Button{
		pin:              pin,
		state:            NeverPressed,
		readings:         NewBuffer(),
		HighMeansPressed: false,
	}
}

// Configure configures the Makey Button pin to have the correct settings to detect touches.
func (b *Button) Configure() error {
	// Note that we have to first turn on the pullup, and then turn off the pullup,
	// in order for the pin to be properly floating.
	b.pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	time.Sleep(10 * time.Millisecond)
	b.pin.Configure(machine.PinConfig{Mode: machine.PinInput})
	b.pin.Set(false)

	return nil
}

// Get returns a ButtonEvent based on the most recent state of the button,
// and if it has changed by being pressed or released.
func (b *Button) Get() ButtonEvent {
	// if pin is pulled up, a low value means the key is pressed
	pressed := !b.pin.Get()
	if b.HighMeansPressed {
		// otherwise, a high value means the key is pressed
		pressed = !pressed
	}

	avg := b.readings.Avg()
	b.readings.Put(pressed)

	switch {
	case pressed && avg > -1*bufferSize+2:
		if b.state == Press {
			return NotChanged
		}

		b.state = Press
		return Pressed
	case !pressed:
		if b.state == Press {
			b.state = Release
			return Released
		}
	}

	return NotChanged
}
