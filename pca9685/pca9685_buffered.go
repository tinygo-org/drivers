package pca9685

import (
	"encoding/binary"

	"tinygo.org/x/drivers"
)

// 16 PWM channels, 2 2 byte values each (on, off 16bits)
const buffLen = 16 * 2 * 2

// DevBuffered provides a way of performing one-shot writes
// on all PWM signals. This is useful when working with systems
// which require as little as possible I/O overhead.
type DevBuffered struct {
	Dev
	// LED buffer, first value is address, following values correspond to LED registers.
	//  [0]: LEDSTART register address
	//  [1:5]: LED0 corresponding to PWM channel 0
	//  [5:9]: LED1 PWM channel 1
	//  ...
	//  [1 + N*4 : 1 + N*4 + 4] : channnel N up to channel 15
	ledBuf [buffLen + 1]byte
}

// New creates a new instance of a PCA9685 device. It performs
// no IO on the i2c bus.
func NewBuffered(bus drivers.I2C, addr uint8) *DevBuffered {
	db := &DevBuffered{
		Dev: New(bus, addr),
	}
	db.ledBuf[0] = LEDSTART
	return db
}

// PrepSet prepares a value to be written to the
// channel's PWM register on Update() call.
func (b *DevBuffered) PrepSet(channel uint8, on uint32) {
	b.PrepPhasedSet(channel, on, 0)
}

// PrepPhasedSet prepares a phased PWM value to be written to the
// channel's register on Update() call.
func (b *DevBuffered) PrepPhasedSet(channel uint8, on, off uint32) {
	onLReg := 1 + channel*4
	binary.LittleEndian.PutUint16(b.ledBuf[onLReg:], uint16(on)&maxtop)
	binary.LittleEndian.PutUint16(b.ledBuf[onLReg+2:], uint16(off)&maxtop)
}

// Update writes the prepared values to the PWM device registers in one shot.
func (b *DevBuffered) Update() error {
	return b.bus.Tx(uint16(b.addr), b.ledBuf[:], nil)
}
