// Package l9110x provides a driver to the L9110/L9110S H-bridge chip
// typically used to control DC motors.
//
// Datasheet: https://www.elecrow.com/download/datasheet-l9110.pdf
//
package l9110x // import "tinygo.org/x/drivers/l9110x"

import (
	"machine"
)

// Device is a motor without speed control.
// ia and ib are the directional pins.
type Device struct {
	ia, ib machine.Pin
}

// New returns a new Motor driver for GPIO-only operation.
func New(direction1, direction2 machine.Pin) Device {
	return Device{
		ia: direction1,
		ib: direction2,
	}
}

// Configure configures the Device.
func (d *Device) Configure() {
	d.ia.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.ib.Configure(machine.PinConfig{Mode: machine.PinOutput})

	d.Stop()
}

// Forward turns motor on in forward direction.
func (d *Device) Forward() {
	d.ia.High()
	d.ib.Low()
}

// Backward turns motor on in backward direction.
func (d *Device) Backward() {
	d.ia.Low()
	d.ib.High()
}

// Stop turns motor off.
func (d *Device) Stop() {
	d.ia.Low()
	d.ib.Low()
}

// PWMDevice is a motor with speed control.
// ia and ib are the directional/speed PWM pins.
type PWMDevice struct {
	ia, ib machine.PWM
}

// NewWithSpeed returns a new PWMMotor driver that uses 2 PWM pins to control both direction and speed.
func NewWithSpeed(direction1, direction2 machine.PWM) PWMDevice {
	return PWMDevice{
		ia: direction1,
		ib: direction2,
	}
}

// Configure configures the PWMDevice.
func (d *PWMDevice) Configure() {
	d.ia.Configure()
	d.ib.Configure()

	d.Stop()
}

// Forward turns motor on in forward direction at specific speed.
func (d *PWMDevice) Forward(speed uint16) {
	d.ia.Set(speed)
	d.ib.Set(0)
}

// Backward turns motor on in backward direction at specific speed.
func (d *PWMDevice) Backward(speed uint16) {
	d.ia.Set(0)
	d.ib.Set(speed)
}

// Stop turns motor off.
func (d *PWMDevice) Stop() {
	d.ia.Set(0)
	d.ib.Set(0)
}
