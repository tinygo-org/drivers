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
	c.Assert(dev.Address, qt.Equals, uint8(ADDRESS))
}

func TestWhoAmI(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, ADDRESS)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.Connected(), qt.Equals, true)

	fake.Registers[WHO_AM_I] = 0x99
	c.Assert(dev.Connected(), qt.Equals, false)
}

// defaultRegisters returns the default values for all of the device's registers.
// see table 22 on page 27 of the datasheet.
func defaultRegisters() []uint8 {
	return []uint8{
		OFFSET_X_REG_L: 0,
		OFFSET_X_REG_H: 0,
		OFFSET_Y_REG_L: 0,
		OFFSET_Y_REG_H: 0,
		OFFSET_Z_REG_L: 0,
		OFFSET_Z_REG_H: 0,
		WHO_AM_I:       0x40,
		CFG_REG_A:      0x03,
		CFG_REG_B:      0,
		CFG_REG_C:      0,
		INT_CRTL_REG:   0xE0,
		INT_SOURCE_REG: 0,
		INT_THS_L_REG:  0,
		INT_THS_H_REG:  0,
		STATUS_REG:     0,
		OUTX_L_REG:     0,
		OUTX_H_REG:     0,
		OUTY_L_REG:     0,
		OUTY_H_REG:     0,
		OUTZ_L_REG:     0,
		OUTZ_H_REG:     0,
		TEMP_OUT_L_REG: 0,
		TEMP_OUT_H_REG: 0,
	}
}
