package seesaw

import "machine"

// assert the machine.I2C conforms to our interface
var _ = I2C(&machine.I2C{})

// I2C represents an I2C bus. It is notably implemented by the
// machine.I2C type.
type I2C interface {
	Tx(addr uint16, w, r []byte) error
}
