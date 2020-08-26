package lis2mdl

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/tester"
)

func TestDefaultI2CAddress(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	dev := New(bus)
	c.Assert(dev.Address, qt.Equals, uint8(MAG_ADDRESS))
}

func TestWhoAmI(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, MAG_ADDRESS)
	bus.AddDevice(fake)

	dev := New(bus)

	fake.SetupRegisters([]uint8{
		0x4F: 0x40,
	})
	c.Assert(dev.Connected(), qt.Equals, true)

	fake.SetupRegister(0x4F, 0x99)
	c.Assert(dev.Connected(), qt.Equals, false)
}
