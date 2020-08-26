package tester

import (
	qt "github.com/frankban/quicktest"
)

// I2CBus implements the I2C interface in memory for testing.
type I2CBus struct {
	C       *qt.C
	Devices []*I2CDevice
}

// NewI2CBus returns an I2CBus mock I2C instance that uses c to flag errors
// if they happen. After creating a I2C instance, add devices
// to it with addDevice before using NewI2CBus interface.
func NewI2CBus(c *qt.C) *I2CBus {
	return &I2CBus{
		C: c,
	}
}

// AddDevice adds a new mock device to the mock I2C bus.
func (bus *I2CBus) AddDevice(d *I2CDevice) {
	bus.Devices = append(bus.Devices, d)
}

// ReadRegister implements I2C.ReadRegister.
func (bus *I2CBus) ReadRegister(addr uint8, r uint8, buf []byte) error {
	return bus.FindDevice(addr).ReadRegister(r, buf)
}

// WriteRegister implements I2C.WriteRegister.
func (bus *I2CBus) WriteRegister(addr uint8, r uint8, buf []byte) error {
	return bus.FindDevice(addr).WriteRegister(r, buf)
}

// Tx implements I2C.Tx.
func (bus *I2CBus) Tx(addr uint16, input, output []byte) error {
	// TODO: implement this
	return nil
}

// FindDevice returns the device with the given address.
func (bus *I2CBus) FindDevice(addr uint8) *I2CDevice {
	for _, dev := range bus.Devices {
		if dev.Addr == addr {
			return dev
		}
	}
	bus.C.Fatalf("invalid device addr %#x passed to i2c bus", addr)
	panic("unreachable")
}
