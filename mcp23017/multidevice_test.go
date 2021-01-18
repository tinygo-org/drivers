package mcp23017

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"tinygo.org/x/drivers/tester"
)

func TestDevicesGetPins(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fdev0 := newDevice(bus, 0x20)
	fdev1 := newDevice(bus, 0x21)
	fdev0.Registers[rGPIO] = 0b10101100
	fdev0.Registers[rGPIO|portB] = 0b01010011
	fdev1.Registers[rGPIO] = 0b10101101
	fdev1.Registers[rGPIO|portB] = 0b01010010
	devs, err := NewI2CDevices(bus, 0x20, 0x21)
	c.Assert(err, qt.IsNil)
	pins := make(PinSlice, 2)
	err = devs.GetPins(pins)
	c.Assert(err, qt.IsNil)
	c.Assert(pins, qt.DeepEquals, PinSlice{0b01010011_10101100, 0b01010010_10101101})

	// It's OK to pass less elements than there are devices.
	pins = make(PinSlice, 1)
	err = devs.GetPins(pins)
	c.Assert(err, qt.IsNil)
	c.Assert(pins, qt.DeepEquals, PinSlice{0b01010011_10101100})
}

func TestDevicesSetPinsAllOff(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fdev0 := newDevice(bus, 0x20)
	fdev1 := newDevice(bus, 0x21)
	fdev0.Registers[rGPIO] = 0b10101100
	fdev0.Registers[rGPIO|portB] = 0b01010011
	fdev1.Registers[rGPIO] = 0b10101101
	fdev1.Registers[rGPIO|portB] = 0b01010010
	devs, err := NewI2CDevices(bus, 0x20, 0x21)
	c.Assert(err, qt.IsNil)

	err = devs.SetPins(nil, PinSlice{0xffff})
	c.Assert(err, qt.IsNil)
	pins := make(PinSlice, 2)
	err = devs.GetPins(pins)
	c.Assert(err, qt.IsNil)
	c.Assert(pins, qt.DeepEquals, PinSlice{0, 0})
}

func TestDevicesSetPinsAllOn(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fdev0 := newDevice(bus, 0x20)
	fdev1 := newDevice(bus, 0x21)
	fdev0.Registers[rGPIO] = 0b10101100
	fdev0.Registers[rGPIO|portB] = 0b01010011
	fdev1.Registers[rGPIO] = 0b10101101
	fdev1.Registers[rGPIO|portB] = 0b01010010
	devs, err := NewI2CDevices(bus, 0x20, 0x21)
	c.Assert(err, qt.IsNil)

	err = devs.SetPins(PinSlice{0xffff}, PinSlice{0xffff})
	c.Assert(err, qt.IsNil)
	pins := make(PinSlice, 2)
	err = devs.GetPins(pins)
	c.Assert(err, qt.IsNil)
	c.Assert(pins, qt.DeepEquals, PinSlice{0xffff, 0xffff})
}

func TestDevicesSetPinsMask(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fdev0 := newDevice(bus, 0x20)
	fdev1 := newDevice(bus, 0x21)
	fdev0.Registers[rGPIO] = 0b10101100
	fdev0.Registers[rGPIO|portB] = 0b01010011
	fdev1.Registers[rGPIO] = 0b10101101
	fdev1.Registers[rGPIO|portB] = 0b01010010
	devs, err := NewI2CDevices(bus, 0x20, 0x21)
	c.Assert(err, qt.IsNil)

	// Sanity check the original value of the pins.
	pins := make(PinSlice, 2)
	err = devs.GetPins(pins)
	c.Assert(err, qt.IsNil)
	c.Assert(pins, qt.DeepEquals, PinSlice{0b01010011_10101100, 0b01010010_10101101})

	pins = make(PinSlice, 2)
	pins.High(0)
	pins.High(1)
	mask := make(PinSlice, 2)
	mask.High(0)
	mask.High(16)

	err = devs.SetPins(pins, mask)
	c.Assert(err, qt.IsNil)
	pins = make(PinSlice, 2)
	err = devs.GetPins(pins)
	c.Assert(err, qt.IsNil)
	c.Assert(pins, qt.DeepEquals, PinSlice{0b01010011_10101101, 0b01010010_10101100})
}

func TestDevicesTogglePins(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	newDevice(bus, 0x20)
	newDevice(bus, 0x21)
	devs, err := NewI2CDevices(bus, 0x20, 0x21)
	c.Assert(err, qt.IsNil)

	mask := make(PinSlice, 2)
	mask.High(0)
	mask.High(16)

	err = devs.TogglePins(mask)
	c.Assert(err, qt.IsNil)
	pins := make(PinSlice, 2)
	err = devs.GetPins(pins)
	c.Assert(err, qt.IsNil)
	c.Assert(pins, qt.DeepEquals, PinSlice{0b00000000_00000001, 0b00000000_00000001})
}

func TestDevicesSetGetModes(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fdev0 := newDevice(bus, 0x20)
	fdev1 := newDevice(bus, 0x21)
	devs, err := NewI2CDevices(bus, 0x20, 0x21)
	c.Assert(err, qt.IsNil)
	// Sanity check that IODIR registers start off all ones.
	c.Assert(fdev0.Registers[rIODIR], qt.Equals, uint8(0xff))

	// The last entry is replicated to fill them all.
	err = devs.SetModes([]PinMode{Input | Pullup, Output})
	c.Assert(err, qt.IsNil)
	c.Assert(fdev0.Registers[rIODIR], qt.Equals, uint8(1))
	c.Assert(fdev0.Registers[rIODIR|portB], qt.Equals, uint8(0))
	c.Assert(fdev1.Registers[rIODIR], qt.Equals, uint8(0))
	c.Assert(fdev1.Registers[rIODIR|portB], qt.Equals, uint8(0))

	modes := make([]PinMode, 2)
	err = devs.GetModes(modes)
	c.Assert(err, qt.Equals, nil)
	c.Assert(modes, qt.DeepEquals, []PinMode{Input | Pullup, Output})

	// It's OK to pass a smaller slice to GetModes.
	modes = make([]PinMode, 1)
	err = devs.GetModes(modes)
	c.Assert(err, qt.Equals, nil)
	c.Assert(modes, qt.DeepEquals, []PinMode{Input | Pullup})
}

func TestDevicesPin(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	newDevice(bus, 0x20)
	fdev1 := newDevice(bus, 0x21)
	devs, err := NewI2CDevices(bus, 0x20, 0x21)
	c.Assert(err, qt.IsNil)
	pin := devs.Pin(16)
	v, err := pin.Get()
	c.Assert(err, qt.Equals, nil)
	c.Assert(v, qt.Equals, false)
	err = pin.High()
	c.Assert(err, qt.Equals, nil)
	c.Assert(fdev1.Registers[rGPIO], qt.Equals, uint8(1))
}

func TestPinSlice(t *testing.T) {
	c := qt.New(t)
	pins := PinSlice(nil).Ensure(20)
	pins.Set(16, true)
	c.Assert(pins, qt.DeepEquals, PinSlice{0, 1})
	pins.Set(31, true)
	c.Assert(pins, qt.DeepEquals, PinSlice{0, 0b10000000_00000001})
	c.Assert(pins.Get(0), qt.Equals, false)
	c.Assert(pins.Get(16), qt.Equals, true)
	pins = pins.Ensure(40)
	c.Assert(pins, qt.DeepEquals, PinSlice{0, 0b10000000_00000001, 0xffff})
	pins.Low(16)
	c.Assert(pins.Get(16), qt.Equals, false)
	pins.High(16)
	c.Assert(pins.Get(16), qt.Equals, true)
	pins.Toggle(16)
	c.Assert(pins.Get(16), qt.Equals, false)
}
