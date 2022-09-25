package tone

import (
	"machine"
)

// PWM is the interface necessary for controlling a speaker.
type PWM interface {
	Configure(config machine.PWMConfig) error
	Channel(pin machine.Pin) (channel uint8, err error)
	Top() uint32
	Set(channel uint8, value uint32)
	SetPeriod(period uint64) error
}

// Speaker is a configured audio output channel based on a PWM.
type Speaker struct {
	pwm PWM
	ch  uint8
}

// New returns a new Speaker instance readily configured for the given PWM and
// pin combination. The lowest frequency possible is 27.5Hz, or A0. The audio
// output uses a PWM so the audio will form a square wave, a sound that
// generally sounds rather harsh.
func New(pwm PWM, pin machine.Pin) (Speaker, error) {
	err := pwm.Configure(machine.PWMConfig{
		Period: uint64(1e9) / 55 / 2,
	})
	if err != nil {
		return Speaker{}, err
	}
	ch, err := pwm.Channel(pin)
	if err != nil {
		return Speaker{}, err
	}
	return Speaker{pwm, ch}, nil
}

// Stop disables the speaker, setting the output to low continuously.
func (s Speaker) Stop() {
	s.pwm.Set(s.ch, 0)
}

// SetPeriod sets the period for the signal in nanoseconds. Use the following
// formula to convert frequency to period:
//
//	period = 1e9 / frequency
//
// You can also use s.SetNote() instead for MIDI note numbers.
func (s Speaker) SetPeriod(period uint64) {
	// Disable output.
	s.Stop()

	if period == 0 {
		// Assume a period of 0 is intended as "no output".
		return
	}

	// Reconfigure period.
	s.pwm.SetPeriod(period)

	// Make this a square wave by setting the channel position to half the
	// period.
	s.pwm.Set(s.ch, s.pwm.Top()/2)
}

// SetNote starts playing the given note. For example, s.SetNote(C4) will
// produce a 440Hz square wave tone.
func (s Speaker) SetNote(note Note) {
	period := note.Period()
	s.SetPeriod(period)
}
