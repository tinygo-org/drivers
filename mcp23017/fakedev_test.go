package mcp23017

import (
	qt "github.com/frankban/quicktest"
)

// fakeBus implements the I2C interface in memory for testing.
type fakeBus struct {
	c    *qt.C
	devs []*fakeDev
}

// newBus returns a fakeBus instance that uses c to flag errors
// if they happen. After creating a fakeBus instance, add devices
// to it with addDevice before using the interface.
func newBus(c *qt.C) *fakeBus {
	return &fakeBus{
		c: c,
	}
}

// fakeDev represents a device on the bus.
type fakeDev struct {
	c    *qt.C
	addr uint8
	// Registers holds the device registers. It can be inspected
	// or changed as desired for testing.
	Registers [registerCount]uint8
	// If Err is non-nil, it will be returned as the error from the
	// I2C methods.
	Err error
}

// addDevice adds a new device at the given address.
func (bus *fakeBus) addDevice(addr uint8) *fakeDev {
	dev := &fakeDev{
		c:    bus.c,
		addr: addr,
		Registers: [registerCount]uint8{
			// IODIRA and IODIRB are all ones by default.
			rIODIR:         0xff,
			rIODIR | portB: 0xff,
		},
	}
	bus.devs = append(bus.devs, dev)
	return dev
}

// ReadRegister implements I2C.ReadRegister.
func (bus *fakeBus) ReadRegister(addr uint8, r uint8, buf []byte) error {
	return bus.findDev(addr).readRegister(r, buf)
}

// WriteRegister implements I2C.WriteRegister.
func (bus *fakeBus) WriteRegister(addr uint8, r uint8, buf []byte) error {
	return bus.findDev(addr).writeRegister(r, buf)
}

func (d *fakeDev) readRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}
	d.assertRegisterRange(r, buf)
	copy(buf, d.Registers[r:])
	return nil
}

func (d *fakeDev) writeRegister(r uint8, buf []byte) error {
	if d.Err != nil {
		return d.Err
	}
	d.assertRegisterRange(r, buf)
	copy(d.Registers[r:], buf)
	return nil
}

// assertRegisterRange asserts that reading or writing the given
// register and subsequent registers is in range of the available registers.
func (d *fakeDev) assertRegisterRange(r uint8, buf []byte) {
	if int(r) >= len(d.Registers) {
		d.c.Fatalf("register read/write [%#x, %#x] start out of range", r, int(r)+len(buf))
	}
	if int(r)+len(buf) > len(d.Registers) {
		d.c.Fatalf("register read/write [%#x, %#x] end out of range", r, int(r)+len(buf))
	}
}

// findDev returns the device with the given address.
func (bus *fakeBus) findDev(addr uint8) *fakeDev {
	for _, dev := range bus.devs {
		if dev.addr == addr {
			return dev
		}
	}
	bus.c.Fatalf("invalid device addr %#x passed to i2c bus", addr)
	panic("unreachable")
}
