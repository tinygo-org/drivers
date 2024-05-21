package servo

import (
	"machine"

	"errors"
)

var ErrInvalidAngle = errors.New("servo: invalid angle")

// PWM is the interface necessary for controlling typical servo motors.
type PWM interface {
	Configure(config machine.PWMConfig) error
	Channel(pin machine.Pin) (channel uint8, err error)
	Top() uint32
	Set(channel uint8, value uint32)
}

// Array is an array of servos controlled by a single PWM peripheral. On most
// chips, one PWM peripheral can control multiple servos (usually two or four).
type Array struct {
	pwm PWM
}

// Servo is a single servo (connected to one PWM output) that's part of a servo
// array.
type Servo struct {
	pwm     PWM
	channel uint8
}

const pwmPeriod = 20e6 // 20ms

// NewArray returns a new servo array based on the given PWM, for if you want to
// control multiple servos from a single PWM peripheral. Using a single PWM for
// multiple servos saves PWM peripherals for other uses and might use less power
// depending on the chip.
//
// If you only want to control a single servo, you could use the New shorthand
// instead.
func NewArray(pwm PWM) (Array, error) {
	err := pwm.Configure(machine.PWMConfig{
		Period: pwmPeriod,
	})
	if err != nil {
		return Array{}, err
	}
	return Array{pwm}, nil
}

// Add adds a new servo to the servo array. Please check the chip documentation
// which pins can be controlled by the given PWM: depending on the chip this
// might be rigid (only a single pin) or very flexible (you can pick any pin).
func (array Array) Add(pin machine.Pin) (Servo, error) {
	channel, err := array.pwm.Channel(pin)
	if err != nil {
		return Servo{}, err
	}
	return Servo{
		pwm:     array.pwm,
		channel: channel,
	}, nil
}

// New is a shorthand for NewArray and array.Add. This is useful if you only
// want to control just a single servo.
func New(pwm PWM, pin machine.Pin) (Servo, error) {
	array, err := NewArray(pwm)
	if err != nil {
		return Servo{}, err
	}
	return array.Add(pin)
}

// SetMicroseconds sets the output signal to be high for the given number of
// microseconds. For many servos the range is normally between 1000µs and 2000µs
// for 90° of rotation (with 1500µs being the 'neutral' middle position).
//
// In many cases they can actually go a bit further, with a wider range of
// supported pulse ranges. For example, they might allow pulse widths from 500µs
// to 2500µs, but be warned that going outside of the 1000µs-2000µs range might
// break the servo as it might destroy the gears if it doesn't support this
// range. Therefore, to be sure check the datasheet before you try values
// outside of the 1000µs-2000µs range.
func (s Servo) SetMicroseconds(microseconds int16) {
	value := uint64(s.pwm.Top()) * uint64(microseconds) / (pwmPeriod / 1000)
	s.pwm.Set(s.channel, uint32(value))
}

// SetAngle sets the angle of the servo in degrees. The angle should be between
// 0 and 180, where 0 is the minimum angle and 180 is the maximum angle.
// This function should work for most servos, but if it doesn't work for yours
// you can use SetMicroseconds directly instead.
func (s Servo) SetAngle(angle int) error {
	if angle < 0 || angle > 180 {
		return ErrInvalidAngle
	}

	// 0° is 1000µs, 180° is 2000µs. See explanation in SetMicroseconds.
	microseconds := angle*1000/180 + 1000
	s.SetMicroseconds(int16(microseconds))

	return nil
}
