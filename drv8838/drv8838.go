// Package drv8838 provides a driver to the DRV8838 Low-Voltage H-Bridge Driver
// typically used to control DC motors.
//
// Datasheet: https://www.ti.com/lit/ds/symlink/drv8838.pdf
//
package drv8838 // import "tinygo.org/x/drivers/drv8838"

import (
	"machine"
)

const (
	DIRECTION_FORWARD  = false
	DIRECTION_BACKWARD = true
)

// Device is a motor with direction (phase), speed (enable) and sleep controlled by a DRV8838
type Device struct {
	enable machine.PWM
	phase  machine.Pin
	sleep  machine.Pin // Optional
}

// Device returns a new motor device
func NewDevice(enable machine.PWM, phase machine.Pin, sleep machine.Pin) Device {
	return Device{
		enable: enable,
		phase:  phase,
		sleep:  sleep,
	}
}

// Configure configures the Device.
func (d *Device) Configure() {
	d.enable.Configure()
	d.phase.Configure(machine.PinConfig{Mode: machine.PinOutput})
	if d.sleep != machine.NoPin {
		d.sleep.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
}

// Forward turns motor on in forward direction at specific speed.
func (d *Device) Forward(speed uint16) {
	d.phase.Set(DIRECTION_FORWARD)
	d.enable.Set(speed)
}

// Backward turns motor on in forward direction at specific speed.
func (d *Device) Backward(speed uint16) {
	d.phase.Set(DIRECTION_BACKWARD)
	d.enable.Set(speed)
}

// Enable sets the speed of the motor
func (d *Device) Enable(speed uint16) {
	d.enable.Set(speed)
}

// Phase sets the direction of the motor
func (d *Device) Phase(direction bool) {
	d.phase.Set(direction)
}

// Sleep sets sleep mode, which allows the motor to coast.
func (d *Device) Sleep(sleep bool) {
	if d.sleep != machine.NoPin {
		d.sleep.Set(sleep)
	}
}

// Stop turns motor off.
func (d *Device) Stop() {
	d.enable.Set(0)
}
