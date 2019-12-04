// Package easystepper provides a simple driver to rotate a 4-wire stepper motor.
package easystepper // import "tinygo.org/x/drivers/easystepper"

import (
	"machine"
	"time"
)

// Device holds the pins and the delay between steps
type Device struct {
	pins       [4]machine.Pin
	stepDelay  int32
	stepNumber int32
}

// DualDevice holds information for controlling 2 motors
type DualDevice struct {
	devices [2]Device
}

// New returns a new easystepper driver given 4 pins numbers (not pin object),
// number of steps and rpm
func New(pin1, pin2, pin3, pin4 machine.Pin, steps int32, rpm int32) Device {
	pin1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pin2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pin3.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pin4.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return Device{
		pins:      [4]machine.Pin{pin1, pin2, pin3, pin4},
		stepDelay: 60000000 / (steps * rpm),
	}
}

// NewDual returns a new dual easystepper driver given 8 pins numbers (not pin object),
// number of steps and rpm
func NewDual(pin1, pin2, pin3, pin4, pin5, pin6, pin7, pin8 machine.Pin, steps int32, rpm int32) DualDevice {
	pin1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pin2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pin3.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pin4.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pin5.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pin6.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pin7.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pin8.Configure(machine.PinConfig{Mode: machine.PinOutput})

	var dual DualDevice
	dual.devices[0] = Device{
		pins:      [4]machine.Pin{pin1, pin2, pin3, pin4},
		stepDelay: 60000000 / (steps * rpm),
	}
	dual.devices[1] = Device{
		pins:      [4]machine.Pin{pin5, pin6, pin7, pin8},
		stepDelay: 60000000 / (steps * rpm),
	}
	return dual
}

// Move rotates the motor the number of given steps
// (negative steps will rotate it the opposite direction)
func (d *Device) Move(steps int32) {
	direction := steps > 0
	if steps < 0 {
		steps = -steps - d.stepNumber
	} else {
		steps += d.stepNumber
	}
	var stepN int8
	var s int32
	d.stepMotor(int8(d.stepNumber % 4))
	for s = d.stepNumber; s < steps; s++ {
		time.Sleep(time.Duration(d.stepDelay) * time.Microsecond)
		if direction {
			stepN = int8(s % 4)
		} else {
			stepN = int8((s + 2*(s%2)) % 4)
		}
		d.stepMotor(stepN)
	}
	d.stepNumber = int32(stepN)
}

// Off turns off all motor pins
func (d *Device) Off() {
	d.pins[0].Low()
	d.pins[1].Low()
	d.pins[2].Low()
	d.pins[3].Low()
}

// Move rotates the motors the number of given steps
// (negative steps will rotate it the opposite direction)
func (d *DualDevice) Move(stepsA, stepsB int32) {
	min := 1
	max := 0
	var directions [2]bool
	var stepN [2]int8
	var minStep int32

	directions[0] = stepsA > 0
	directions[1] = stepsB > 0
	if stepsA < 0 {
		stepsA = -stepsA
	}
	if stepsB < 0 {
		stepsB = -stepsB
	}

	if stepsB > stepsA {
		stepsA, stepsB = stepsB, stepsA
		min = 0
		max = 1
	}
	d.devices[0].stepMotor(int8(d.devices[0].stepNumber % 4))
	d.devices[1].stepMotor(int8(d.devices[1].stepNumber % 4))
	for s := int32(0); s < stepsA; s++ {
		time.Sleep(time.Duration(d.devices[0].stepDelay) * time.Microsecond)
		if directions[max] {
			stepN[max] = int8(s % 4)
		} else {
			stepN[max] = int8((s + 2*(s%2)) % 4)
		}

		if ((s*stepsB)/stepsA) > minStep {
			minStep++
			if directions[min] {
				stepN[min] = int8(minStep % 4)
			} else {
				stepN[min] = int8((minStep + 2*(minStep%2)) % 4)
			}
		}

		d.devices[0].stepMotor(stepN[0])
		d.devices[1].stepMotor(stepN[1])
	}

}

// Off turns off all motor pins
func (d *DualDevice) Off() {
	d.devices[0].Off()
	d.devices[1].Off()
}

// stepMotor changes the pins' state to the correct step
func (d *Device) stepMotor(step int8) {
	switch step {
	case 0:
		d.pins[0].High()
		d.pins[1].Low()
		d.pins[2].High()
		d.pins[3].Low()
		break
	case 1:
		d.pins[0].Low()
		d.pins[1].High()
		d.pins[2].High()
		d.pins[3].Low()
		break
	case 2:
		d.pins[0].Low()
		d.pins[1].High()
		d.pins[2].Low()
		d.pins[3].High()
		break
	case 3:
		d.pins[0].High()
		d.pins[1].Low()
		d.pins[2].Low()
		d.pins[3].High()
		break
	}
}
