// Simple driver to rotate a 4-wire stepper motor
package easystepper

import (
	"machine"
	"time"
)

// Device holds the pins and the delay between steps
type Device struct {
	pins       [4]machine.GPIO
	stepDelay  int32
	stepNumber int32
}

// New returns a new easystepper driver given 4 pins numbers (not pin object),
// number of steps and rpm
func New(pin1 uint8, pin2 uint8, pin3 uint8, pin4 uint8, steps int32, rpm int32) Device {
	m1 := machine.GPIO{pin1}
	m1.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	m2 := machine.GPIO{pin2}
	m2.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	m3 := machine.GPIO{pin3}
	m3.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	m4 := machine.GPIO{pin4}
	m4.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	return Device{
		pins:      [4]machine.GPIO{m1, m2, m3, m4},
		stepDelay: 60000000 / (steps * rpm),
	}
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
