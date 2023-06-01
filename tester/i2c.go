package tester

import "fmt"

// I2CBus implements the I2C interface in memory for testing.
type I2CBus struct {
	c       Failer
	devices []I2CDevice
}

// NewI2CBus returns an I2CBus mock I2C instance that uses c to flag errors
// if they happen. After creating a I2C instance, add devices
// to it with addDevice before using NewI2CBus interface.
func NewI2CBus(c Failer) *I2CBus {
	return &I2CBus{
		c: c,
	}
}

// AddDevice adds a new mock device to the mock I2C bus.
// It panics if a device with the same address is added more than once.
func (bus *I2CBus) AddDevice(d I2CDevice) {
	for _, dev := range bus.devices {
		if dev.Addr() == d.Addr() {
			panic(fmt.Errorf("device already added at address %#x", d))
		}
	}
	bus.devices = append(bus.devices, d)
}

// NewDevice creates a new device with the given address
// and adds it to the mock I2C bus.
func (bus *I2CBus) NewDevice(addr uint8) *I2CDevice8 {
	dev := NewI2CDevice8(bus.c, addr)
	bus.AddDevice(dev)
	return dev
}

// ReadRegister implements I2C.ReadRegister.
func (bus *I2CBus) ReadRegister(addr uint8, r uint8, buf []byte) error {
	return bus.FindDevice(addr).readRegister(r, buf)
}

// WriteRegister implements I2C.WriteRegister.
func (bus *I2CBus) WriteRegister(addr uint8, r uint8, buf []byte) error {
	return bus.FindDevice(addr).writeRegister(r, buf)
}

// Tx implements I2C.Tx.
func (bus *I2CBus) Tx(addr uint16, w, r []byte) error {
	return bus.FindDevice(uint8(addr)).Tx(w, r)
}

// FindDevice returns the device with the given address.
func (bus *I2CBus) FindDevice(addr uint8) I2CDevice {
	for _, dev := range bus.devices {
		if dev.Addr() == addr {
			return dev
		}
	}
	bus.c.Fatalf("invalid device addr %#x passed to i2c bus", addr)
	panic("unreachable")
}
