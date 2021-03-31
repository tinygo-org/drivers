package tester

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestCreate16(t *testing.T) {
	c := qt.New(t)
	bus := NewI2CBus(c)
	d := NewI2CDevice16(c, 8)
	bus.AddDevice(d)
}

func TestRead16(t *testing.T) {
	c := qt.New(t)
	bus := NewI2CBus(c)
	d := NewI2CDevice16(c, 8)
	bus.AddDevice(d)

	// Setup a random register
	d.Registers[3] = 0x1234

	buf := []byte{0, 0}
	err := bus.ReadRegister(8, 3, buf)
	c.Assert(err, qt.IsNil)
	c.Assert(buf[0], qt.Equals, byte(0x12))
	c.Assert(buf[1], qt.Equals, byte(0x34))
}

func TestWrite16(t *testing.T) {
	c := qt.New(t)
	bus := NewI2CBus(c)
	d := NewI2CDevice16(c, 8)
	bus.AddDevice(d)

	d.Registers[9] = 0x0
	err := bus.WriteRegister(8, 9, []byte{0xbe, 0xad})
	c.Assert(err, qt.IsNil)
	c.Assert(d.Registers[9], qt.Equals, uint16(0xbead))
}
