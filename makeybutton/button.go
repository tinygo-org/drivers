// Package makeybutton providers a driver for a button that can be triggered
// by anything that is conductive by using an ultra high value resistor.
//
// Inspired by the amazing MakeyMakey
// https://makeymakey.com/
//
// This code is a reinterpretation of
// https://github.com/sparkfun/MaKeyMaKey/blob/master/firmware/Arduino/makey_makey/makey_makey.ino
package makeybutton

import (
	"machine"
	"time"
)

var (
	pressThreshold   int = 1
	releaseThreshold int = 0
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
	pressed          bool
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
	// Note that on AVR we have to first turn on the pullup, and then turn off the pullup,
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
	b.update()

	if b.pressed {
		// the button had previously been pressed,
		// but now appears to have been released.
		if b.readings.Sum() <= releaseThreshold {
			b.pressed = false
			b.state = Release
			return Released
		}
	} else {
		// the button had previously not been pressed,
		// but now appears to have been pressed.
		if b.readings.Sum() >= pressThreshold {
			b.pressed = true
			b.state = Press
			return Pressed
		}
	}

	return NotChanged
}

func (b *Button) update() {
	// if pin is pulled up, a low value means the key is pressed
	press := !b.pin.Get()
	if b.HighMeansPressed {
		// otherwise, a high value means the key is pressed
		press = !press
	}

	b.readings.Put(press)
}
