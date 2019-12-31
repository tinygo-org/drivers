// Package l293x provides a driver to the L293/L293D H-bridge chip
// typically used to control DC motors.
//
// Datasheet: https://www.ti.com/lit/ds/symlink/l293d.pdf
//
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

// PWMDevice is a motor with speed control.
// a1 and a2 are the directional GPIO pins.
// en is the PWM pin that controls the motor speed.
type PWMDevice struct {
	a1, a2 machine.Pin
	en     machine.PWM
}

// NewWithSpeed returns a new PWMMotor driver that uses a PWM pin to control speed.
func NewWithSpeed(direction1, direction2 machine.Pin, speedPin machine.PWM) PWMDevice {
	return PWMDevice{
		a1: direction1,
		a2: direction2,
		en: speedPin,
	}
}

// Configure configures the PWMDevice.
func (d *PWMDevice) Configure() {
	d.a1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.a2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.en.Configure()

	d.Stop()
}

// Forward turns motor on in forward direction at specific speed.
func (d *PWMDevice) Forward(speed uint16) {
	d.a1.High()
	d.a2.Low()
	d.en.Set(speed)
}

// Backward turns motor on in backward direction at specific speed.
func (d *PWMDevice) Backward(speed uint16) {
	d.a1.Low()
	d.a2.High()
	d.en.Set(speed)
}

// Stop turns motor off.
func (d *PWMDevice) Stop() {
	d.a1.Low()
	d.a2.Low()
	d.en.Set(0)
}
