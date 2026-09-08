// Package accel implemeents methods for detecting orientation of the HUB75 RGB
// LED matrix using an ST LIS3DH 3-axis linear accelerometer connected via I2C
// interface.
package accel

import (
	"errors"
	"strconv"
	"strings"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/lis3dh"
)

// ErrNotConnected indicates I2C communication failed with the LIS3DH.
var ErrNotConnected = errors.New("not connected to accelerometer")

// Accel encapsulates an I2C connection with the LIS3DH.
type Accel struct {
	*lis3dh.Device
	connected bool
}

// Config defines the configuration parameters for the acceleration sensor.
type Config struct {
	Address uint16
	Range   lis3dh.Range
}

// New creates and returns a new Accel. The Configure method must be called on
// the returned object before the sensor can be used.
//
// The given I2C interface must be configured for use prior to calling New.
func New(i2c drivers.I2C) *Accel {
	dev := lis3dh.New(i2c)
	return &Accel{Device: &dev}
}

// Configure initializes the LIS3DH sensor with given configuration. This method
// must be called before the sensor can be used.
func (acc *Accel) Configure(config Config) error {
	acc.Device.Address = config.Address
	acc.Device.Configure()
	acc.Device.SetRange(config.Range)
	acc.connected = acc.Connected()
	if !acc.connected {
		return ErrNotConnected
	}
	return nil
}

// Get returns the acceleration due to gravity detected in each of the three
// dimensions (x, y, and z).
func (acc *Accel) Get() (x, y, z int, err error) {
	if acc.connected {
		ax, ay, az, err := acc.ReadAcceleration()
		return int(ax), int(ay), int(az), err
	}
	return 0, 0, 0, ErrNotConnected
}

// String constructs a string representation of the acceleration due to gravity
// detected in each of the three dimensions (x, y, and z).
func (acc *Accel) String() string {
	if x, y, z, err := acc.Get(); nil == err {
		xs := strconv.FormatInt(int64(x), 10)
		ys := strconv.FormatInt(int64(y), 10)
		zs := strconv.FormatInt(int64(z), 10)
		return strings.Join([]string{xs, ys, zs}, ", ")
	}
	return ""
}
