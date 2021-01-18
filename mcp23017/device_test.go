package mcp23017

import (
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"

	"tinygo.org/x/drivers/tester"
)

func TestGetPins(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fdev := newDevice(bus, 0x20)
	fdev.Registers[rGPIO] = 0b10101100
	fdev.Registers[rGPIO|portB] = 0b01010011
	dev, err := NewI2C(bus, 0x20)
	c.Assert(err, qt.IsNil)
	pins, err := dev.GetPins()
	c.Assert(err, qt.IsNil)
	c.Assert(pins, qt.Equals, Pins(0b01010011_10101100))
}

func TestSetPins(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fdev := newDevice(bus, 0x20)
	fdev.Registers[rGPIO] = 0b00001111
	fdev.Registers[rGPIO|portB] = 0b11110000
	dev, err := NewI2C(bus, 0x20)
	c.Assert(err, qt.IsNil)
	pins, err := dev.GetPins()
	c.Assert(err, qt.IsNil)
	c.Assert(pins, qt.Equals, Pins(0b11110000_00001111))

	err = dev.SetPins(0b01100000_00110000, 0b10101010_01010101)
	c.Assert(err, qt.IsNil)
	pins, err = dev.GetPins()
	c.Assert(err, qt.IsNil)
	c.Assert(pins, qt.Equals, Pins(0b01110000_0001_1010))

	// The logic uses the cached value of the pins rather than
	// reading it from the registers each time.
	fdev.Registers[rGPIO] = 0
	fdev.Registers[rGPIO|portB] = 0

	err = dev.SetPins(0b01000000_00110000, 0b01100000_00000000)
	c.Assert(err, qt.IsNil)
	pins, err = dev.GetPins()
	c.Assert(err, qt.IsNil)
	c.Assert(pins, qt.Equals, Pins(0b01010000_00011010))
}

func TestTogglePins(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fdev := newDevice(bus, 0x20)
	fdev.Registers[rGPIO] = 0b00001111
	fdev.Registers[rGPIO|portB] = 0b11110000
	dev, err := NewI2C(bus, 0x20)
	c.Assert(err, qt.IsNil)
	pins, err := dev.GetPins()
	c.Assert(err, qt.IsNil)
	c.Assert(pins, qt.Equals, Pins(0b11110000_00001111))

	err = dev.TogglePins(0b10101010_01010101)
	c.Assert(err, qt.IsNil)
	pins, err = dev.GetPins()
	c.Assert(err, qt.IsNil)
	c.Assert(pins, qt.Equals, Pins(0b01011010_01011010))
}

func TestSetGetModes(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fdev := newDevice(bus, 0x20)
	dev, err := NewI2C(bus, 0x20)
	c.Assert(err, qt.IsNil)
	// Calling SetModes with less items in than there are
	// pins should use the last item for all the unspecified ones.
	err = dev.SetModes([]PinMode{Input | Invert, Output})
	c.Assert(err, qt.IsNil)
	c.Assert(fdev.Registers[rIODIR], qt.Equals, uint8(0b00000001))
	c.Assert(fdev.Registers[rIOPOL], qt.Equals, uint8(0b00000001))
	c.Assert(fdev.Registers[rGPPU], qt.Equals, uint8(0))

	modes := make([]PinMode, 17)
	err = dev.GetModes(modes)
	c.Assert(err, qt.IsNil)
	c.Assert(modes[0], qt.Equals, Input|Invert)
	for i, m := range modes[1:16] {
		c.Assert(m, qt.Equals, Output, qt.Commentf("index %d", i))
	}
	c.Assert(modes[16], qt.Equals, PinMode(0))

	// Using an empty slice should reset all the modes to the initial state.
	err = dev.SetModes(nil)
	c.Assert(err, qt.IsNil)
	c.Assert(fdev.Registers[rIODIR], qt.Equals, uint8(0b11111111))
	c.Assert(fdev.Registers[rIOPOL], qt.Equals, uint8(0))
	c.Assert(fdev.Registers[rGPPU], qt.Equals, uint8(0))
}

func TestPinSetGet(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fdev := newDevice(bus, 0x20)
	dev, err := NewI2C(bus, 0x20)
	c.Assert(err, qt.IsNil)
	pin := dev.Pin(1)
	v, err := pin.Get()
	c.Assert(err, qt.Equals, nil)
	c.Assert(v, qt.Equals, false)
	err = pin.Set(true)
	c.Assert(err, qt.Equals, nil)
	c.Assert(fdev.Registers[rGPIO], qt.Equals, uint8(0b10))
	v, err = pin.Get()
	c.Assert(err, qt.Equals, nil)
	c.Assert(v, qt.Equals, true)
	err = pin.Set(false)
	c.Assert(err, qt.Equals, nil)
	c.Assert(fdev.Registers[rGPIO], qt.Equals, uint8(0))
	err = pin.High()
	c.Assert(err, qt.Equals, nil)
	c.Assert(fdev.Registers[rGPIO], qt.Equals, uint8(0b10))
	err = pin.Low()
	c.Assert(err, qt.Equals, nil)
	c.Assert(fdev.Registers[rGPIO], qt.Equals, uint8(0))
}

func TestPinToggle(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fdev := newDevice(bus, 0x20)
	dev, err := NewI2C(bus, 0x20)
	c.Assert(err, qt.IsNil)
	pin := dev.Pin(1)
	v, err := pin.Get()
	c.Assert(err, qt.Equals, nil)
	c.Assert(v, qt.Equals, false)
	err = pin.Toggle()
	c.Assert(err, qt.Equals, nil)
	c.Assert(fdev.Registers[rGPIO], qt.Equals, uint8(0b10))
	err = pin.Toggle()
	c.Assert(err, qt.Equals, nil)
	c.Assert(fdev.Registers[rGPIO], qt.Equals, uint8(0))
}

func TestPinMode(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fdev := newDevice(bus, 0x20)
	dev, err := NewI2C(bus, 0x20)
	c.Assert(err, qt.IsNil)
	pin := dev.Pin(1)
	mode, err := pin.GetMode()
	c.Assert(err, qt.IsNil)
	c.Assert(mode, qt.Equals, PinMode(0))
	c.Assert(mode&Direction, qt.Equals, Input)

	err = pin.SetMode(Input | Pullup | Invert)
	c.Assert(err, qt.IsNil)
	c.Assert(fdev.Registers[rIODIR], qt.Equals, uint8(0b11111111))
	c.Assert(fdev.Registers[rIOPOL], qt.Equals, uint8(0b10))
	c.Assert(fdev.Registers[rGPPU], qt.Equals, uint8(0b10))

	mode, err = pin.GetMode()
	c.Assert(err, qt.IsNil)
	c.Assert(mode, qt.Equals, Input|Pullup|Invert)

	// Set another pin to output.
	err = dev.Pin(2).SetMode(Output)
	c.Assert(err, qt.IsNil)
	c.Assert(fdev.Registers[rIODIR], qt.Equals, uint8(0b11111011))
	c.Assert(fdev.Registers[rIOPOL], qt.Equals, uint8(0b10))
	c.Assert(fdev.Registers[rGPPU], qt.Equals, uint8(0b10))

	// Check that changing a pin in port B works too.
	err = dev.Pin(8).SetMode(Output)
	c.Assert(err, qt.IsNil)
	c.Assert(fdev.Registers[rIODIR], qt.Equals, uint8(0b11111011))
	c.Assert(fdev.Registers[rIODIR|portB], qt.Equals, uint8(0b11111110))
	c.Assert(fdev.Registers[rIOPOL], qt.Equals, uint8(0b10))
	c.Assert(fdev.Registers[rIOPOL|portB], qt.Equals, uint8(0))
	c.Assert(fdev.Registers[rGPPU], qt.Equals, uint8(0b10))
	c.Assert(fdev.Registers[rGPPU|portB], qt.Equals, uint8(0))
}

func TestPins(t *testing.T) {
	c := qt.New(t)
	var p Pins
	p.Set(1, true)
	c.Assert(p, qt.Equals, Pins(0b10))
	c.Assert(p.Get(1), qt.Equals, true)
	c.Assert(p.Get(0), qt.Equals, false)
	c.Assert(p.Get(16), qt.Equals, false)
	p.High(2)
	c.Assert(p, qt.Equals, Pins(0b110))
	p.Low(1)
	c.Assert(p, qt.Equals, Pins(0b100))
}

func TestInitWithError(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fdev := newDevice(bus, 0x20)
	fdev.Err = fmt.Errorf("some error")
	dev, err := NewI2C(bus, 0x20)
	c.Assert(err, qt.ErrorMatches, `cannot initialize mcp23017 device at 0x20: some error`)
	c.Assert(dev, qt.IsNil)
}

func newDevice(bus *tester.I2CBus, addr uint8) *tester.I2CDevice {
	fdev := bus.NewDevice(addr)
	// IODIRA and IODIRB are all ones by default.
	fdev.Registers[rIODIR] = 0xff
	fdev.Registers[rIODIR|portB] = 0xff
	return fdev
}
