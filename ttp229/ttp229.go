// Package ttp229 is for the 16 keys or 8 keys touch pad detector IC
// Datasheet (BSF version): https://www.sunrom.com/download/SUNROM-TTP229-BSF_V1.1_EN.pdf
package ttp229 // import "tinygo.org/x/drivers/ttp229"

import (
	"time"

	"machine"
)

// Device wraps a connection to a TTP229 device.
type Device struct {
	bus      Buser
	keys     uint16
	prevKeys uint16
	inputs   byte
}

// PinBus holds the structure for GPIO protocol
type PinBus struct {
	scl machine.Pin
	sdo machine.Pin
}

// Buser interface since there are different versions with different protocols
type Buser interface {
	readBits(data byte) uint16
}

// Configuration
type Configuration struct {
	Inputs byte
}

// NewPin creates a new TTP229 connection through 2 machine.Pin, this is suitable for the BSF variant of the TTP229.
func NewPin(scl, sdo machine.Pin) Device {
	scl.Configure(machine.PinConfig{Mode: machine.PinOutput})
	sdo.Configure(machine.PinConfig{Mode: machine.PinInput})
	return Device{
		bus: &PinBus{
			scl: scl,
			sdo: sdo,
		},
		inputs: 16,
	}
}

// Configure sets up the device for communication
func (d *Device) Configure(cfg Configuration) bool {
	if cfg.Inputs != 0 {
		d.inputs = cfg.Inputs
	}
	return true
}

// ReadKeys returns the pressed keys as bits of a uint16
func (d *Device) ReadKeys() uint16 {
	d.prevKeys = d.keys
	d.keys = d.bus.readBits(d.inputs)
	return d.keys
}

// bitRead return if the specific bit is on/off of the given uint16
func (d *Device) bitRead(number uint16, bit byte) bool {
	return (number & (0x0001 << bit)) != 0
}

// GetKey returns the current pressed key (only returns one key, even in multitouch mode)
func (d *Device) GetKey() int8 {
	for i := byte(0); i < d.inputs; i++ {
		if !d.bitRead(d.keys, byte(i)) {
			return int8(i)
		}
	}
	return -1
}

// IsKeyPressed returns true if the given key is pressed
func (d *Device) IsKeyPressed(key byte) bool {
	return !d.bitRead(d.keys, key)
}

// IsKeyDown returns true if the given key was just pressed (and it wasn't previously)
func (d *Device) IsKeyDown(key byte) bool {
	return d.bitRead(d.prevKeys, key) && !d.bitRead(d.keys, key)
}

// IsKeyUp returns true if the given key was released
func (d *Device) IsKeyUp(key byte) bool {
	return !d.bitRead(d.prevKeys, key) && d.bitRead(d.keys, key)
}

// readBits returns the pressed keys as bits of a uint16
func (b *PinBus) readBits(inputs byte) uint16 {
	b.scl.High()
	time.Sleep(1 * time.Millisecond)
	var pressed uint16
	for i := byte(0); i < inputs; i++ {
		b.scl.Low()
		time.Sleep(1 * time.Millisecond)
		if b.sdo.Get() {
			pressed = pressed | (0x0001 << i)
		}
		b.scl.High()
		time.Sleep(1 * time.Millisecond)
	}
	time.Sleep(1 * time.Millisecond)
	return pressed
}
