// Package easystepper provides a simple driver to rotate a 4-wire stepper motor.
package easystepper // import "tinygo.org/x/drivers/easystepper"

import (
	"errors"
	"machine"
	"time"
)

// StepMode determines the coil sequence used to perform a single step
type StepMode uint8

// Valid values for StepMode
const (
	// ModeFour uses a 'four step' coil sequence (12-23-34-41). This is the default (zero-value) mode
	ModeFour StepMode = iota
	// ModeEight uses an 'eight step' coil sequence (1-12-2-23-3-34-4-41)
	ModeEight
)

// stepCount is a helper function to return the number of steps in a StepMode sequence
func (sm StepMode) stepCount() uint {
	switch sm {
	default:
		fallthrough
	case ModeFour:
		return 4
	case ModeEight:
		return 8
	}
}

// DeviceConfig contains the configuration data for a single easystepper driver
type DeviceConfig struct {
	// Pin1 ... Pin4 determines the pins to configure and use for the device
	Pin1, Pin2, Pin3, Pin4 machine.Pin
	// StepCount is the number of steps required to perform a full revolution of the stepper motor
	StepCount uint
	// RPM determines the speed of the stepper motor in 'Revolutions per Minute'
	RPM uint
	// Mode determines the coil sequence used to perform a single step
	Mode StepMode
}

// DualDeviceConfig contains the configuration data for a dual easystepper driver
type DualDeviceConfig struct {
	DeviceConfig
	// Pin5 ... Pin8 determines the pins to configure and use for the second device
	Pin5, Pin6, Pin7, Pin8 machine.Pin
}

// Device holds the pins and the delay between steps
type Device struct {
	pins       [4]machine.Pin
	stepDelay  time.Duration
	stepNumber uint8
	stepMode   StepMode
}

// DualDevice holds information for controlling 2 motors
type DualDevice struct {
	devices [2]*Device
}

// New returns a new single easystepper driver given a DeviceConfig
func New(config DeviceConfig) (*Device, error) {
	if config.StepCount == 0 || config.RPM == 0 {
		return nil, errors.New("config.StepCount and config.RPM must be > 0")
	}
	return &Device{
		pins:      [4]machine.Pin{config.Pin1, config.Pin2, config.Pin3, config.Pin4},
		stepDelay: time.Second * 60 / time.Duration((config.StepCount * config.RPM)),
		stepMode:  config.Mode,
	}, nil
}

// Configure configures the pins of the Device
func (d *Device) Configure() {
	for _, pin := range d.pins {
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
}

// NewDual returns a new dual easystepper driver given 8 pins, number of steps and rpm
func NewDual(config DualDeviceConfig) (*DualDevice, error) {
	// Create the first device
	dev1, err := New(config.DeviceConfig)
	if err != nil {
		return nil, err
	}
	// Create the second device
	config.DeviceConfig.Pin1 = config.Pin5
	config.DeviceConfig.Pin2 = config.Pin6
	config.DeviceConfig.Pin3 = config.Pin7
	config.DeviceConfig.Pin4 = config.Pin8
	dev2, err := New(config.DeviceConfig)
	if err != nil {
		return nil, err
	}
	// Return composite dual device
	return &DualDevice{devices: [2]*Device{dev1, dev2}}, nil
}

// Configure configures the pins of the DualDevice
func (d *DualDevice) Configure() {
	d.devices[0].Configure()
	d.devices[1].Configure()
}

// Move rotates the motor the number of given steps
// (negative steps will rotate it the opposite direction)
func (d *Device) Move(steps int32) {
	direction := steps > 0
	if steps < 0 {
		steps = -steps
	}
	steps += int32(d.stepNumber)
	var s int32
	d.stepMotor(d.stepNumber)
	for s = int32(d.stepNumber); s < steps; s++ {
		time.Sleep(d.stepDelay)
		d.moveDirectionSteps(direction, s)
	}
}

// Off turns off all motor pins
func (d *Device) Off() {
	for _, pin := range d.pins {
		pin.Low()
	}
}

// Move rotates the motors the number of given steps
// (negative steps will rotate it the opposite direction)
func (d *DualDevice) Move(stepsA, stepsB int32) {
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
	d.devices[0].stepMotor(d.devices[0].stepNumber)
	d.devices[1].stepMotor(d.devices[1].stepNumber)
	stepsA += int32(d.devices[max].stepNumber)
	minStep = int32(d.devices[min].stepNumber)
	for s := int32(d.devices[max].stepNumber); s < stepsA; s++ {
		time.Sleep(d.devices[0].stepDelay)
		d.devices[max].moveDirectionSteps(directions[max], s)

		if ((s * stepsB) / stepsA) > minStep {
			minStep++
			d.devices[min].moveDirectionSteps(directions[min], minStep)
		}
	}
}

// Off turns off all motor pins
func (d *DualDevice) Off() {
	d.devices[0].Off()
	d.devices[1].Off()
}

// stepMotor changes the pins' state to the correct step
func (d *Device) stepMotor(step uint8) {
	switch d.stepMode {
	default:
		fallthrough
	case ModeFour:
		d.stepMotor4(step)
	case ModeEight:
		d.stepMotor8(step)
	}
}

// stepMotor4 changes the pins' state to the correct step in 4-step mode
func (d *Device) stepMotor4(step uint8) {
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
func (d *Device) stepMotor8(step uint8) {
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
func (d *Device) moveDirectionSteps(direction bool, step int32) {
	modulus := int32(d.stepMode.stepCount())
	if direction {
		d.stepMotor(uint8(step % modulus))
	} else {
		d.stepMotor(uint8(((-step % modulus) + modulus) % modulus))
	}
}
