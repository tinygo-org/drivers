// Package l9110x provides a driver to the L9110/L9110S H-bridge chip
// typically used to control DC motors.
//
// Datasheet: https://www.elecrow.com/download/datasheet-l9110.pdf
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

// PWM is the interface necessary for controlling the motor driver.
type PWM interface {
	Configure(config machine.PWMConfig) error
	Channel(pin machine.Pin) (channel uint8, err error)
	Top() uint32
	Set(channel uint8, value uint32)
	SetPeriod(period uint64) error
}

// PWMDevice is a motor with speed control.
// ia and ib are the directional/speed PWM pins.
type PWMDevice struct {
	pwm    PWM
	ca, cb uint8
}

// NewWithSpeed returns a new PWMMotor driver that uses 2 PWM pins to control both direction and speed.
func NewWithSpeed(ca, cb uint8, pwm PWM) PWMDevice {
	return PWMDevice{
		pwm: pwm,
		ca:  ca,
		cb:  cb,
	}
}

// Configure configures the PWMDevice. Note that the pins, PWM interface,
// and channels must all already be configured.
func (d *PWMDevice) Configure() (err error) {
	d.Stop()
	return
}

// Forward turns motor on in forward direction at specific speed as a percentage.
func (d *PWMDevice) Forward(speed uint32) {
	d.pwm.Set(d.ca, d.pwm.Top()*speed/100)
	d.pwm.Set(d.cb, 0)
}

// Backward turns motor on in backward direction at specific speed as a percentage.
func (d *PWMDevice) Backward(speed uint32) {
	d.pwm.Set(d.ca, 0)
	d.pwm.Set(d.cb, d.pwm.Top()*speed/100)
}

// Stop turns motor off.
func (d *PWMDevice) Stop() {
	d.pwm.Set(d.ca, 0)
	d.pwm.Set(d.cb, 0)
}
