package tester

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestCreate8(t *testing.T) {
	c := qt.New(t)
	bus := NewI2CBus(c)
	d := NewI2CDevice8(c, 8)
	bus.AddDevice(d)
}

func TestRead8(t *testing.T) {
	c := qt.New(t)
	bus := NewI2CBus(c)
	d := NewI2CDevice8(c, 8)
	bus.AddDevice(d)

	// Setup a random register
	d.Registers[3] = 0x12
	d.Registers[4] = 0x34

	buf := []byte{0, 0}
	err := bus.ReadRegister(8, 3, buf)
	c.Assert(err, qt.IsNil)
	c.Assert(buf[0], qt.Equals, byte(0x12))
	c.Assert(buf[1], qt.Equals, byte(0x34))
}

func TestWrite8(t *testing.T) {
	c := qt.New(t)
	bus := NewI2CBus(c)
	d := NewI2CDevice8(c, 8)
	bus.AddDevice(d)

	err := bus.WriteRegister(8, 9, []byte{0xbe, 0xad})
	c.Assert(err, qt.IsNil)
	c.Assert(d.Registers[9], qt.Equals, uint8(0xbe))
	c.Assert(d.Registers[10], qt.Equals, uint8(0xad))
}
