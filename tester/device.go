package tester

import (
	qt "github.com/frankban/quicktest"
)

const maxRegisters = 200

// I2CDevice represents a mock I2C device on a mock I2C bus.
type I2CDevice struct {
	C    *qt.C
	Addr uint8
	// Registers holds the device registers. It can be inspected
	// or changed as desired for testing.
	Registers [maxRegisters]uint8
	// If Err is non-nil, it will be returned as the error from the
	// I2C methods.
	Err error
}

// NewI2CDevice returns a new mock I2C device.
func NewI2CDevice(c *qt.C, addr uint8) *I2CDevice {
	return &I2CDevice{
		C:    c,
		Addr: addr,
	}
}

// ReadRegister implements I2C.ReadRegister.
func (d *I2CDevice) ReadRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}
	d.AssertRegisterRange(r, buf)
	copy(buf, d.Registers[r:])
	return nil
}

// WriteRegister implements I2C.WriteRegister.
func (d *I2CDevice) WriteRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}
	d.AssertRegisterRange(r, buf)
	copy(d.Registers[r:], buf)
	return nil
}

// AssertRegisterRange asserts that reading or writing the given
// register and subsequent registers is in range of the available registers.
func (d *I2CDevice) AssertRegisterRange(r uint8, buf []byte) {
	if int(r) >= len(d.Registers) {
		d.C.Fatalf("register read/write [%#x, %#x] start out of range", r, int(r)+len(buf))
	}
	if int(r)+len(buf) > len(d.Registers) {
		d.C.Fatalf("register read/write [%#x, %#x] end out of range", r, int(r)+len(buf))
	}
}
