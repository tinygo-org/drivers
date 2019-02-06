// Package bh1750 provides a driver for the BH1750 digital Ambient Light
//
// Datasheet:
// https://www.mouser.com/ds/2/348/bh1750fvi-e-186247.pdf
//
package bh1750

import (
	"time"

	"machine"
)

// Device wraps an I2C connection to a bh1750 device.
type Device struct {
	bus  machine.I2C
	mode byte
}

// New creates a new bh1750 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus machine.I2C) Device {
	return Device{
		bus:  bus,
		mode: CONTINUOUS_HIGH_RES_MODE,
	}
}

// Configure sets up the device for communication
func (d *Device) Configure() {
	d.bus.Tx(Address, []byte{POWER_ON}, nil)
	d.SetMode(d.mode)
}

// RawSensorData returns the raw value from the bh1750
func (d *Device) RawSensorData() uint16 {

	buf := []byte{1, 0}
	d.bus.Tx(Address, nil, buf)
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
	// 1.2 = measurement accuracy as per the datasheet
	return int32(float32(100*coef*lux) / 1.2)
}

// SetMode changes the reading mode for the sensor
func (d *Device) SetMode(mode byte) {
	if mode == CONTINUOUS_HIGH_RES_MODE ||
		mode == CONTINUOUS_HIGH_RES_MODE_2 ||
		mode == CONTINUOUS_LOW_RES_MODE ||
		mode == ONE_TIME_HIGH_RES_MODE ||
		mode == ONE_TIME_HIGH_RES_MODE_2 ||
		mode == ONE_TIME_LOW_RES_MODE {
			d.mode = mode
			d.bus.Tx(Address, []byte{d.mode}, nil)
			time.Sleep(10*time.Millisecond)
	}
}
