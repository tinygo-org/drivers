package tester

import (
	"errors"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestReadAddr16(t *testing.T) {
	c := qt.New(t)
	bus := NewI2CBus(c)
	d := NewI2CDeviceAddr16(c, 0x29)
	bus.AddDevice(d)

	d.Registers[0x010F] = 0xEA
	d.Registers[0x0110] = 0xAA

	buf := []byte{0, 0}
	err := bus.Tx(0x29, []byte{0x01, 0x0F}, buf)
	c.Assert(err, qt.IsNil)
	c.Assert(buf, qt.DeepEquals, []byte{0xEA, 0xAA})
}

func TestWriteAddr16(t *testing.T) {
	c := qt.New(t)
	bus := NewI2CBus(c)
	d := NewI2CDeviceAddr16(c, 0x29)
	bus.AddDevice(d)

	err := bus.Tx(0x29, []byte{0x04, 0x76, 0xbe, 0xad}, nil)
	c.Assert(err, qt.IsNil)
	c.Assert(d.Registers[0x0476], qt.Equals, byte(0xbe))
	c.Assert(d.Registers[0x0477], qt.Equals, byte(0xad))
}

func TestErrAddr16(t *testing.T) {
	c := qt.New(t)
	bus := NewI2CBus(c)
	d := NewI2CDeviceAddr16(c, 0x29)
	bus.AddDevice(d)

	d.Err = errors.New("bus error")
	err := bus.Tx(0x29, []byte{0x00, 0x00}, make([]byte, 1))
	c.Assert(err, qt.Equals, d.Err)
}
