// Package buzzer provides a very simplistic driver for a connected buzzer or low-fidelity speaker.
package buzzer // import "tinygo.org/x/drivers/buzzer"

import (
	"machine"

	"time"
)

// Device wraps a GPIO connection to a buzzer.
type Device struct {
	pin  machine.Pin
	High bool
	BPM  float64
}

// New returns a new buzzer driver given which pin to use
func New(pin machine.Pin) Device {
	return Device{
		pin:  pin,
		High: false,
		BPM:  96.0,
	}
}

// On sets the buzzer to a high state.
func (l *Device) On() (err error) {
	l.pin.Set(true)
	l.High = true
	return
}

// Off sets the buzzer to a low state.
func (l *Device) Off() (err error) {
	l.pin.Set(false)
	l.High = false
	return
}

// Toggle sets the buzzer to the opposite of it's current state
func (l *Device) Toggle() (err error) {
	if l.High {
		err = l.Off()
	} else {
		err = l.On()
	}
	return
}

// Tone plays a tone of the requested frequency and duration.
func (l *Device) Tone(hz, duration float64) (err error) {
	// calculation based off https://www.arduino.cc/en/Tutorial/Melody
	tone := (1.0 / (2.0 * hz)) * 1000000.0

	tempo := ((60 / l.BPM) * (duration * 1000))

	// no tone during rest, just let the duration pass.
	if hz == Rest {
		time.Sleep(time.Duration(tempo) * time.Millisecond)
		return
	}

	for i := 0.0; i < tempo*1000; i += tone * 2.0 {
		if err = l.On(); err != nil {
			return
		}
		time.Sleep(time.Duration(tone) * time.Microsecond)

		if err = l.Off(); err != nil {
			return
		}
		time.Sleep(time.Duration(tone) * time.Microsecond)
	}

	return
}
