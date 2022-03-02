// Package easystepper provides a simple driver to rotate a 4-wire stepper motor.
package easystepper // import "tinygo.org/x/drivers/easystepper"

import (
	"machine"
	"time"
)

// StepMode determines the number of coil configurations for a single step
type StepMode uint8

// Valid values for StepMode
const (
	FourStepMode  = 4 // 12-23-34-41
	EightStepMode = 8 // 1-12-2-23-3-34-4-41
)

// Device defines the interface for a single easystepper driver
type Device interface {
	// Configure configures the pins of the Device
	Configure()
	// Move rotates the motor the given number of steps (negative steps rotates in opposite direction)
	// This method uses the '4-Step' model 12-23-34-41 and is retained for backwards compatibility
	Move(steps int32)
	// MoveStepsMode rotates the motor the given number of steps using the given step mode
	// Negative steps rotates in opposite direction)
	MoveStepsMode(steps int32, mode StepMode)
	// Off turns off all motor pins
	Off()
}

// DualDevice defines the interface for a dual easystepper driver
type DualDevice interface {
	// Configure configures the pins of the DualDevice
	Configure()
	// Move rotates the motor the given number of steps (negative steps rotates in opposite direction)
	// This method uses the '4-Step' model 12-23-34-41 and is retained for backwards compatibility
	Move(stepsA, stepsB int32)
	// MoveStepsMode rotates the motors the given number of steps using the given step mode
	// Negative steps rotates in opposite direction)
	MoveStepsMode(stepsA, stepsB int32, mode StepMode)
	// Off turns off all motor pins
	Off()
}

// Device holds the pins and the delay between steps
type device struct {
	pins       [4]machine.Pin
	stepDelay  int32
	stepNumber uint8
}

// DualDevice holds information for controlling 2 motors
type dualDevice struct {
	devices [2]device
}

// New returns a new easystepper driver given 4 pins, number of steps and rpm
func New(pin1, pin2, pin3, pin4 machine.Pin, steps int32, rpm int32) Device {
	return &device{
		pins:      [4]machine.Pin{pin1, pin2, pin3, pin4},
		stepDelay: 60000000 / (steps * rpm),
	}
}

// Configure configures the pins of the Device
func (d *device) Configure() {
	for _, pin := range d.pins {
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
}

// NewDual returns a new dual easystepper driver given 8 pins, number of steps and rpm
func NewDual(pin1, pin2, pin3, pin4, pin5, pin6, pin7, pin8 machine.Pin, steps int32, rpm int32) DualDevice {
	var dual dualDevice
	dual.devices[0] = device{
		pins:      [4]machine.Pin{pin1, pin2, pin3, pin4},
		stepDelay: 60000000 / (steps * rpm),
	}
	dual.devices[1] = device{
		pins:      [4]machine.Pin{pin5, pin6, pin7, pin8},
		stepDelay: 60000000 / (steps * rpm),
	}
	return &dual
}

// Configure configures the pins of the DualDevice
func (d *dualDevice) Configure() {
	d.devices[0].Configure()
	d.devices[1].Configure()
}

// Move rotates the motor the number of given steps using 4-step mode
// (negative steps will rotate it the opposite direction)
func (d *device) Move(steps int32) {
	d.MoveStepsMode(steps, FourStepMode)
}

// MoveStepsMode rotates the motor the number of given steps using the given step mode
// (negative steps will rotate it the opposite direction)
func (d *device) MoveStepsMode(steps int32, mode StepMode) {
	direction := steps > 0
	if steps < 0 {
		steps = -steps
	}
	steps += int32(d.stepNumber)
	var s int32
	d.stepMotor(d.stepNumber, mode)
	for s = int32(d.stepNumber); s < steps; s++ {
		time.Sleep(time.Duration(d.stepDelay) * time.Microsecond)
		d.moveDirectionSteps(direction, s, mode)
	}
}

// Off turns off all motor pins
func (d *device) Off() {
	for _, pin := range d.pins {
		pin.Low()
	}
}

// Move rotates the motor the number of given steps using 4-step mode
// (negative steps will rotate it the opposite direction)
func (d *dualDevice) Move(stepsA, stepsB int32) {
	d.MoveStepsMode(stepsA, stepsB, FourStepMode)
}

// MoveStepsMode rotates the motor the number of given steps using the given step mode
// (negative steps will rotate it the opposite direction)
func (d *dualDevice) MoveStepsMode(stepsA, stepsB int32, mode StepMode) {
	min := uint8(1)
	max := uint8(0)
	var directions [2]bool
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
		max, min = min, max
	}
	d.devices[0].stepMotor(d.devices[0].stepNumber, mode)
	d.devices[1].stepMotor(d.devices[1].stepNumber, mode)
	stepsA += int32(d.devices[max].stepNumber)
	minStep = int32(d.devices[min].stepNumber)
	for s := int32(d.devices[max].stepNumber); s < stepsA; s++ {
		time.Sleep(time.Duration(d.devices[0].stepDelay) * time.Microsecond)
		d.devices[max].moveDirectionSteps(directions[max], s, mode)

		if ((s * stepsB) / stepsA) > minStep {
			minStep++
			d.devices[min].moveDirectionSteps(directions[min], minStep, mode)
		}
	}
}

// Off turns off all motor pins
func (d *dualDevice) Off() {
	d.devices[0].Off()
	d.devices[1].Off()
}

// stepMotor changes the pins' state to the correct step for the given mode
func (d *device) stepMotor(step uint8, mode StepMode) {
	if mode == FourStepMode {
		d.stepMotor4(step)
	} else if mode == EightStepMode {
		d.stepMotor8(step)
	}
}

// stepMotor4 changes the pins' state to the correct step in 4-step mode
func (d *device) stepMotor4(step uint8) {
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
	d.stepNumber = step
}

// stepMotor8 changes the pins' state to the correct step in 8-step mode
func (d *device) stepMotor8(step uint8) {
	switch step {
	case 0:
		d.pins[0].High()
		d.pins[2].Low()
		d.pins[1].Low()
		d.pins[3].Low()
	case 1:
		d.pins[0].High()
		d.pins[2].High()
		d.pins[1].Low()
		d.pins[3].Low()
	case 2:
		d.pins[0].Low()
		d.pins[2].High()
		d.pins[1].Low()
		d.pins[3].Low()
	case 3:
		d.pins[0].Low()
		d.pins[2].High()
		d.pins[1].High()
		d.pins[3].Low()
	case 4:
		d.pins[0].Low()
		d.pins[2].Low()
		d.pins[1].High()
		d.pins[3].Low()
	case 5:
		d.pins[0].Low()
		d.pins[2].Low()
		d.pins[1].High()
		d.pins[3].High()
	case 6:
		d.pins[0].Low()
		d.pins[2].Low()
		d.pins[1].Low()
		d.pins[3].High()
	case 7:
		d.pins[0].High()
		d.pins[2].Low()
		d.pins[1].Low()
		d.pins[3].High()
	}
	d.stepNumber = step
}

// moveDirectionSteps uses the direction to calculate the correct step and change the motor to it.
// Direction true:  (4-step mode) 0, 1, 2, 3, 0, 1, 2, ...
// Direction false: (4-step mode) 0, 3, 2, 1, 0, 3, 2, ...
// Direction true:  (8-step mode) 0, 1, 2, 3, 4, 5, 6, 7, 0, 1, 2, ...
// Direction false: (8-step mode) 0, 7, 6, 5, 4, 3, 2, 1, 0, 7, 6, ...
func (d *device) moveDirectionSteps(direction bool, step int32, mode StepMode) {
	modulus := int32(mode)
	if direction {
		d.stepMotor(uint8(step%modulus), mode)
	} else {
		d.stepMotor(uint8(((-step%modulus)+modulus)%modulus), mode)
	}
}
