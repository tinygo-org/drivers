// Package l293x provides a driver to the L293/L293D H-bridge chip
// typically used to control DC motors.
//
// Datasheet: https://www.ti.com/lit/ds/symlink/l293d.pdf
package l293x // import "tinygo.org/x/drivers/l293x"

import (
	"machine"
)

// Device is a motor without speed control.
// a1 and a2 are the directional pins.
// en is the pin turns the motor on/off.
type Device struct {
	a1, a2 machine.Pin
	en     machine.Pin
}

// New returns a new Motor driver for GPIO-only operation.
func New(direction1, direction2, enablePin machine.Pin) Device {
	return Device{
		a1: direction1,
		a2: direction2,
		en: enablePin,
	}
}

// Configure configures the Device.
func (d *Device) Configure() {
	d.a1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.a2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.en.Configure(machine.PinConfig{Mode: machine.PinOutput})

	d.Stop()
}

// Forward turns motor on in forward direction.
func (d *Device) Forward() {
	d.a1.High()
	d.a2.Low()
	d.en.High()
}

// Backward turns motor on in backward direction.
func (d *Device) Backward() {
	d.a1.Low()
	d.a2.High()
	d.en.High()
}

// Stop turns motor off.
func (d *Device) Stop() {
	d.a1.Low()
	d.a2.Low()
	d.en.Low()
}

// PWM is the interface necessary for controlling the motor driver.
type PWM interface {
	Configure(config machine.PWMConfig) error
	Channel(pin machine.Pin) (channel uint8, err error)
	Top() uint32
	Set(channel uint8, value uint32)
	SetPeriod(period uint64) error
}

// PWMDevice is a motor with speed control.
// a1 and a2 are the directional GPIO pins.
// en is the PWM pin that controls the motor speed.
type PWMDevice struct {
	a1, a2 machine.Pin
	spc    uint8
	pwm    PWM
}

// NewWithSpeed returns a new PWMMotor driver that uses an already configured PWM channel
// to control speed.
func NewWithSpeed(direction1, direction2 machine.Pin, spc uint8, pwm PWM) PWMDevice {
	return PWMDevice{
		a1:  direction1,
		a2:  direction2,
		spc: spc,
		pwm: pwm,
	}
}

// Configure configures the PWMDevice. Note that the PWM interface and
// channel must already be configured, this function will not do it for you.
func (d *PWMDevice) Configure() error {
	d.a1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.a2.Configure(machine.PinConfig{Mode: machine.PinOutput})

	d.Stop()

	return nil
}

// Forward turns motor on in forward direction at specific speed as a percentage.
func (d *PWMDevice) Forward(speed uint32) {
	if speed > 100 {
		speed = 100
	}

	d.a1.High()
	d.a2.Low()
	d.pwm.Set(d.spc, d.pwm.Top()*speed/100)
}

// Backward turns motor on in backward direction at specific speed as a percentage.
func (d *PWMDevice) Backward(speed uint32) {
	if speed > 100 {
		speed = 100
	}

	d.a1.Low()
	d.a2.High()
	d.pwm.Set(d.spc, d.pwm.Top()*speed/100)
}

// Stop turns motor off.
func (d *PWMDevice) Stop() {
	d.a1.Low()
	d.a2.Low()
	d.pwm.Set(d.spc, 0)
}
