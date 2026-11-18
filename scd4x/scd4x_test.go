package scd4x

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/tester"
)

func TestDefaultI2CAddress(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	dev := New(bus)
	c.Assert(dev.Address, qt.Equals, uint8(Address))
}

func TestCRC8(t *testing.T) {
	c := qt.New(t)
	c.Assert(crc8([]byte{0xbe, 0xef}), qt.Equals, uint8(0x92))
}

func TestConversions(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	dev := New(bus)
	dev.temperature = 0xffff
	dev.humidity = 0xffff

	c.Assert(dev.Temperature(), qt.Equals, int32(129997))
	c.Assert(dev.Humidity(), qt.Equals, int32(9999))
}

func TestUpdateKeepsCachedDataWhenNotReady(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDeviceCmd(c, Address)
	ready := &tester.Cmd{
		Command:  []byte{0xe4, 0xb8},
		Mask:     []byte{0xff, 0xff},
		Response: []byte{0x00, 0x00, 0x81},
	}
	fake.Commands = map[uint8]*tester.Cmd{0: ready}
	bus.AddDevice(fake)
	dev := New(bus)
	dev.co2 = 420
	dev.temperature = 1000
	dev.humidity = 2000

	c.Assert(dev.Update(drivers.Concentration), qt.IsNil)
	c.Assert(ready.Invocations, qt.Equals, 1)
	c.Assert(dev.co2, qt.Equals, uint16(420))
	c.Assert(dev.temperature, qt.Equals, uint16(1000))
	c.Assert(dev.humidity, qt.Equals, uint16(2000))
}

func TestUpdateReadsDataAndChecksCRC(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDeviceCmd(c, Address)
	ready := &tester.Cmd{
		Command:  []byte{0xe4, 0xb8},
		Mask:     []byte{0xff, 0xff},
		Response: []byte{0x00, 0x01, 0xb0},
	}
	measurement := &tester.Cmd{
		Command:  []byte{0xec, 0x05},
		Mask:     []byte{0xff, 0xff},
		Response: []byte{0x01, 0xa4, 0x00, 0x66, 0x66, 0x00, 0x80, 0x00, 0x00},
	}
	for i := 0; i < len(measurement.Response); i += 3 {
		measurement.Response[i+2] = crc8(measurement.Response[i : i+2])
	}
	fake.Commands = map[uint8]*tester.Cmd{0: ready, 1: measurement}
	bus.AddDevice(fake)
	dev := New(bus)

	c.Assert(dev.Update(drivers.Temperature), qt.IsNil)
	c.Assert(measurement.Invocations, qt.Equals, 1)
	c.Assert(dev.co2, qt.Equals, uint16(420))
	c.Assert(dev.temperature, qt.Equals, uint16(0x6666))
	c.Assert(dev.humidity, qt.Equals, uint16(0x8000))

	measurement.Response[8] ^= 0xff
	c.Assert(dev.Update(drivers.Humidity), qt.Equals, ErrCRC)
	c.Assert(dev.co2, qt.Equals, uint16(420))
	c.Assert(dev.temperature, qt.Equals, uint16(0x6666))
	c.Assert(dev.humidity, qt.Equals, uint16(0x8000))
}
