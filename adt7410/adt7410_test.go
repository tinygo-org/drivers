package adt7410

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/tester"
)

func TestDefaultI2CAddress(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	dev := New(bus)
	c.Assert(dev.Address, qt.Equals, uint8(Address))
}

func TestWhoAmI(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, Address)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.Connected(), qt.Equals, true)

	fake.Registers[RegID] = 0x99
	c.Assert(dev.Connected(), qt.Equals, false)
}

// defaultRegisters returns the default values for all of the device's registers.
// see table 22 on page 27 of the datasheet.
func defaultRegisters() []uint8 {
	return []uint8{
		RegTempValueMSB: 0,
		RegTempValueLSB: 0,
		RegStatus:       0,
		RegConfig:       0,
		RegTHIGHMsbReg:  0x20,
		RegTHIGHLsbReg:  0,
		RegTLOWMsbReg:   0x05,
		RegTLOWLsbReg:   0,
		RegTCRITMsbReg:  0x49,
		RegTCRITLsbReg:  0x80,
		RegTHYSTReg:     0x05,
		RegID:           0xC8,
		RegReset:        0,
	}
}
