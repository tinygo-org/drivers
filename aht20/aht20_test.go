package aht20

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/tester"
)

func TestDefaultI2CAddress(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	dev := New(bus)
	c.Assert(uint8(dev.Address), qt.Equals, uint8(Address))
}

func TestInitialization(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fdev := tester.NewI2CDeviceCmd(c, Address)
	fdev.Commands = defaultCommands()
	bus.AddDevice(fdev)

	// Set status to uninitialized to force initialization
	fdev.Commands[CMD_STATUS].Response[0] = 0x0C

	dev := New(bus)
	dev.Configure()

	// Check initialization command invoked
	c.Assert(fdev.Commands[CMD_INITIALIZE].Invocations > 0, qt.Equals, true)
}

func TestRead(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fdev := tester.NewI2CDeviceCmd(c, Address)
	fdev.Commands = defaultCommands()
	bus.AddDevice(fdev)

	dev := New(bus)
	dev.Read()

	// Should be 25deg (250 decidegrees)
	c.Assert(dev.DeciCelsius(), qt.Equals, int32(250))

	// Should be 36.3% (363 decipercent)
	c.Assert(dev.DeciRelHumidity(), qt.Equals, int32(363))
}

func defaultCommands() map[uint8]*tester.Cmd {
	return map[uint8]*tester.Cmd{
		CMD_INITIALIZE: {
			Command:  []byte{0xBE},
			Mask:     []byte{0xFF},
			Response: []byte{},
		},
		CMD_TRIGGER: {
			Command:  []byte{0xAC, 0x33, 0x00},
			Mask:     []byte{0xFF, 0xFF, 0xFF},
			Response: []byte{0x1C, 0x5D, 0x10, 0x66, 0x01, 0xD2, 0x93},
		},
		CMD_SOFTRESET: {
			Command:  []byte{0xBA},
			Mask:     []byte{0xFF},
			Response: []byte{},
		},
		CMD_STATUS: {
			Command:  []byte{0x71},
			Mask:     []byte{0xFF},
			Response: []byte{0x1C},
		},
	}
}
