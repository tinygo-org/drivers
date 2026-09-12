// Package easystepper provides a simple driver to rotate a 4-wire stepper motor.
package easystepper // import "tinygo.org/x/drivers/easystepper"

import (
	"errors"
	"machine"
	"time"
)

var (
	ErrRPM = errors.New("rpm must be greater than zero")
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

	// stepCount is the number of steps for one full revolution.
	// SetRPM needs it to calculate a new stepDelay.
	stepCount uint

	// remainingSteps is how many steps MoveAsync has left to do.
	remainingSteps uint32

	// direction is true to move forward and false to move backward.
	direction bool

	// moving is true when the motor has movement scheduled.
	moving bool

	// continuous is true to move until Stop. If it is false, movement
	// stops when remainingSteps is zero.
	continuous bool

	// nextStep is the time of the next step. Update uses it to return
	// immediately when no step is necessary.
	nextStep time.Time
}

// DualDevice holds information for controlling 2 motors
type DualDevice struct {
	devices [2]*Device

	// moving is true when a non blocking dual movement is active.
	moving bool

	// continuous is true to move both motors until Stop.
	continuous bool

	// directions is the direction of each motor.
	directions [2]bool

	// totalSteps is the number of steps requested for each motor.
	// Update needs it to divide the steps of the slower motor.
	totalSteps [2]uint32

	// completedSteps is the number of steps each motor has done.
	completedSteps [2]uint32

	// primary is the motor with the most steps. It controls the
	// timing of a coordinated MoveAsync.
	primary uint8

	// secondary is the other motor.
	secondary uint8

	// nextStep is the time of the next coordinated step.
	nextStep time.Time
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
		stepCount: config.StepCount,
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

// Move rotates the motor the number of given steps and waits until the
// movement is complete. Negative steps rotate it the opposite direction.
func (d *Device) Move(steps int32) {
	if steps == 0 {
		return
	}
	direction := steps > 0
	if steps < 0 {
		steps = -steps
	}
	for i := int32(0); i < steps; i++ {
		d.step(direction)
		time.Sleep(d.stepDelay)
	}
}

// MoveAsync schedules a number of steps and returns immediately. Negative
// steps move backward. You must call Update to make the motor move.
func (d *Device) MoveAsync(steps int32) {
	if steps == 0 {
		return
	}
	d.continuous = false
	d.direction = steps > 0
	if steps < 0 {
		steps = -steps
	}
	d.remainingSteps = uint32(steps)
	d.moving = true
	d.nextStep = time.Now()
}

// Start moves the motor until Stop and returns immediately. You must call
// Update to make the motor move.
func (d *Device) Start(direction bool) {
	d.direction = direction
	d.continuous = true
	d.moving = true
	d.nextStep = time.Now()
}

// Update does one step if a step is due. It does not block. Call it
// frequently from the main loop.
func (d *Device) Update() {
	if !d.moving {
		return
	}
	now := time.Now()
	if now.Before(d.nextStep) {
		return
	}
	d.step(d.direction)
	if !d.continuous {
		d.remainingSteps--
		if d.remainingSteps == 0 {
			d.moving = false
			return
		}
	}
	d.nextStep = schedule(d.nextStep, now, d.stepDelay)
}

// schedule gives the time of the next step. It adds the delay to the last
// time to prevent drift, but starts from now if the caller is very late.
func schedule(last, now time.Time, delay time.Duration) time.Time {
	next := last.Add(delay)
	if next.Before(now) {
		return now.Add(delay)
	}
	return next
}

// Stop ends the movement from MoveAsync or Start. The coils stay on, so the
// motor holds its position. Use Off to also remove power from the coils.
func (d *Device) Stop() {
	d.moving = false
	d.continuous = false
	d.remainingSteps = 0
}

// IsMoving tells you if the motor has an active movement.
func (d *Device) IsMoving() bool {
	return d.moving
}

// Off turns off all motor pins. This removes power from the coils, so the
// motor does not hold its position.
func (d *Device) Off() {
	for _, pin := range d.pins {
		pin.Low()
	}
}

// SetRPM changes the speed of the motor. You can call it while the motor
// moves. The new speed applies to the steps that follow.
func (d *Device) SetRPM(rpm uint) error {
	if rpm == 0 {
		return ErrRPM
	}
	d.stepDelay = time.Second * 60 / time.Duration(d.stepCount*rpm)
	return nil
}

// Move rotates the motors the number of given steps
// (negative steps will rotate it the opposite direction)
func (d *DualDevice) Move(stepsA, stepsB int32) {
	if stepsA == 0 && stepsB == 0 {
		return
	}
	primary, secondary, directions, totals := d.plan(stepsA, stepsB)
	var completed [2]uint32

	for completed[primary] < totals[primary] {
		d.devices[primary].step(directions[primary])
		completed[primary]++

		if completed[secondary] < share(completed[primary], totals, primary, secondary) {
			d.devices[secondary].step(directions[secondary])
			completed[secondary]++
		}
		time.Sleep(d.devices[primary].stepDelay)
	}
}

// plan gives the motor with the most steps, the other motor, the direction
// of each motor, and the number of steps each motor must do.
func (d *DualDevice) plan(stepsA, stepsB int32) (uint8, uint8, [2]bool, [2]uint32) {
	directions := [2]bool{stepsA > 0, stepsB > 0}
	if stepsA < 0 {
		stepsA = -stepsA
	}
	if stepsB < 0 {
		stepsB = -stepsB
	}
	primary, secondary := uint8(0), uint8(1)
	if stepsB > stepsA {
		primary, secondary = 1, 0
	}
	return primary, secondary, directions, [2]uint32{uint32(stepsA), uint32(stepsB)}
}

// share gives the steps the slower motor must have done. It needs 64 bits
// because two step counts near 65535 overflow a uint32.
func share(done uint32, totals [2]uint32, primary, secondary uint8) uint32 {
	return uint32(uint64(done) * uint64(totals[secondary]) / uint64(totals[primary]))
}

// MoveAsync starts a movement of both motors and returns immediately. Both
// motors stop together, as with Move. Call Update to make them move.
func (d *DualDevice) MoveAsync(stepsA, stepsB int32) {
	if stepsA == 0 && stepsB == 0 {
		return
	}
	// Update drives each motor directly in this mode, so cancel any
	// movement that Start gave to the two motors.
	d.devices[0].Stop()
	d.devices[1].Stop()

	d.primary, d.secondary, d.directions, d.totalSteps = d.plan(stepsA, stepsB)
	d.completedSteps[0] = 0
	d.completedSteps[1] = 0
	d.continuous = false
	d.moving = true
	d.nextStep = time.Now()
}

// SetRPM changes the speed of both motors.
func (d *DualDevice) SetRPM(rpm uint) error {
	return d.SetRPMs(rpm, rpm)
}

// SetRPMs changes the speed of each motor. Different speeds turn a robot
// that has one motor on each wheel.
func (d *DualDevice) SetRPMs(rpmA, rpmB uint) error {
	if err := d.devices[0].SetRPM(rpmA); err != nil {
		return err
	}
	return d.devices[1].SetRPM(rpmB)
}

// Update does the next step of a DualDevice movement if a step is due. It
// does not block. Call it frequently from the main loop.
func (d *DualDevice) Update() {
	if !d.moving {
		return
	}

	// After Start each motor keeps its own speed and its own timing.
	if d.continuous {
		d.devices[0].Update()
		d.devices[1].Update()
		return
	}

	now := time.Now()
	if now.Before(d.nextStep) {
		return
	}

	primary := d.primary
	secondary := d.secondary

	d.devices[primary].step(d.directions[primary])
	d.completedSteps[primary]++

	if d.completedSteps[secondary] < share(d.completedSteps[primary], d.totalSteps, primary, secondary) {
		d.devices[secondary].step(d.directions[secondary])
		d.completedSteps[secondary]++
	}

	if d.completedSteps[primary] >= d.totalSteps[primary] {
		d.moving = false
		return
	}

	d.nextStep = schedule(d.nextStep, now, d.devices[primary].stepDelay)
}

// Start moves both motors until Stop and returns immediately. Each motor
// keeps its own speed from SetRPMs. Call Update to make them move.
func (d *DualDevice) Start(directionA, directionB bool) {
	d.continuous = true
	d.moving = true
	d.devices[0].Start(directionA)
	d.devices[1].Start(directionB)
}

// Stop ends all movement. The coils stay on, so the motors hold their
// position. Use Off to also remove power from the coils.
func (d *DualDevice) Stop() {
	d.moving = false
	d.continuous = false
	d.devices[0].Stop()
	d.devices[1].Stop()
}

// IsMoving tells you if one of the motors has an active movement.
func (d *DualDevice) IsMoving() bool {
	if d.continuous {
		return d.devices[0].IsMoving() || d.devices[1].IsMoving()
	}
	return d.moving
}

// Off turns off all motor pins
func (d *DualDevice) Off() {
	d.devices[0].Off()
	d.devices[1].Off()
}

// step moves the motor one step. It does not wait, so Move and Update can
// both use it. Forward in 4 step mode gives 0, 1, 2, 3, 0, 1, and backward
// gives 0, 3, 2, 1, 0, 3.
func (d *Device) step(direction bool) {
	// Length of the coil sequence, which is 4 or 8. This is not the
	// stepCount field, which is the steps for one revolution.
	seq := uint8(d.stepMode.stepCount())
	if direction {
		d.stepNumber = (d.stepNumber + 1) % seq
	} else {
		d.stepNumber = (d.stepNumber + seq - 1) % seq
	}
	d.stepMotor(d.stepNumber)
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
