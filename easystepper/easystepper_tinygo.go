//go:build baremetal

package easystepper

import (
	"errors"
	"machine"
	"time"

	"tinygo.org/x/drivers/internal/pin"
)

// New returns a new single easystepper driver given a DeviceConfig
func New(config DeviceConfig) (*Device, error) {
	if config.StepCount == 0 || config.RPM == 0 {
		return nil, errors.New("config.StepCount and config.RPM must be > 0")
	}
	return &Device{
		pins:      [4]pin.OutputFunc{config.Pin1.Set, config.Pin2.Set, config.Pin3.Set, config.Pin4.Set},
		stepDelay: time.Second * 60 / time.Duration((config.StepCount * config.RPM)),
		stepMode:  config.Mode,
		config: func() {
			pin.ConfigureOutput(config.Pin1)
			pin.ConfigureOutput(config.Pin2)
			pin.ConfigureOutput(config.Pin3)
			pin.ConfigureOutput(config.Pin4)
		},
	}, nil
}

// Configure configures the pins of the Device
func (d *Device) Configure() {
	if d.config == nil {
		panic(pin.ErrConfigBeforeInstantiated)
	}
	d.config()
}

// Configure configures the pins of the DualDevice
func (d *DualDevice) Configure() {
	d.devices[0].Configure()
	d.devices[1].Configure()
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
