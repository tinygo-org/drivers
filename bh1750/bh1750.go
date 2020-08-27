// Package bh1750 provides a driver for the BH1750 digital Ambient Light
//
// Datasheet:
// https://www.mouser.com/ds/2/348/bh1750fvi-e-186247.pdf
//
package bh1750 // import "tinygo.org/x/drivers/bh1750"

import (
	"time"

	"tinygo.org/x/drivers"
)

// SamplingMode is the sampling's resolution of the measurement
type SamplingMode byte

// Device wraps an I2C connection to a bh1750 device.
type Device struct {
	bus     drivers.I2C
	Address uint16
	mode    SamplingMode
}

// New creates a new bh1750 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
		mode:    CONTINUOUS_HIGH_RES_MODE,
	}
}

// Configure sets up the device for communication
func (d *Device) Configure() {
	d.bus.Tx(d.Address, []byte{POWER_ON}, nil)
	d.SetMode(d.mode)
}

// RawSensorData returns the raw value from the bh1750
func (d *Device) RawSensorData() uint16 {

	buf := []byte{1, 0}
	d.bus.Tx(d.Address, nil, buf)
	return (uint16(buf[0]) << 8) | uint16(buf[1])
}

// Illuminance returns the adjusted value in mlx (milliLux)
func (d *Device) Illuminance() int32 {

	lux := uint32(d.RawSensorData())
	var coef uint32
	if d.mode == CONTINUOUS_HIGH_RES_MODE || d.mode == ONE_TIME_HIGH_RES_MODE {
		coef = HIGH_RES
	} else if d.mode == CONTINUOUS_HIGH_RES_MODE_2 || d.mode == ONE_TIME_HIGH_RES_MODE_2 {
		coef = HIGH_RES2
	} else {
		coef = LOW_RES
	}
	// 100 * coef * lux * (5/6)
	// 5/6 = measurement accuracy as per the datasheet
	return int32(250 * coef * lux / 3)
}

// SetMode changes the reading mode for the sensor
func (d *Device) SetMode(mode SamplingMode) {
	d.mode = mode
	d.bus.Tx(d.Address, []byte{byte(d.mode)}, nil)
	time.Sleep(10 * time.Millisecond)
}
