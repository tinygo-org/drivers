package as7262

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/tester"
)

func TestDefaultI2CAddress(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	dev := New(bus)
	c.Assert(dev.Address, qt.Equals, uint8(DefaultAddress))
}

func TestWhoAmI(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.Connected(), qt.Equals, true)

	fake.Registers[HardwareVersionReg] = 0xFF
	c.Assert(dev.Connected(), qt.Equals, false)
}

func TestReadViolet(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.ReadViolet(), qt.Equals, float32(0.15625))

	fake.Registers[VCalReg+1] = byte(0b00111111)
	c.Assert(dev.ReadViolet(), qt.Not(qt.Equals), float32(0.15625))

}

func TestReadBlue(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.ReadBlue(), qt.Equals, float32(0.22222))

	fake.Registers[BCalReg+1] = byte(0b00111111)
	c.Assert(dev.ReadBlue(), qt.Not(qt.Equals), float32(0.22222))

}

func TestReadGreen(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.ReadGreen(), qt.Equals, float32(1.45))

	fake.Registers[GCalReg+1] = byte(0b00000001)
	c.Assert(dev.ReadGreen(), qt.Not(qt.Equals), float32(1.45))

}

func TestReadYellow(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.ReadYellow(), qt.Equals, float32(0.00002))

	fake.Registers[YCalReg+1] = byte(0b00000001)
	c.Assert(dev.ReadYellow(), qt.Not(qt.Equals), float32(0.00002))

}

func TestReadOrange(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.ReadOrange(), qt.Equals, float32(0.15625))

	fake.Registers[OCalReg+1] = byte(0b00000001)
	c.Assert(dev.ReadOrange(), qt.Not(qt.Equals), float32(0.15625))

}

func TestReadRed(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.ReadRed(), qt.Equals, float32(0.15625))

	fake.Registers[RCalReg+1] = byte(0b00000001)
	c.Assert(dev.ReadRed(), qt.Not(qt.Equals), float32(0.15625))

}

func TestReadRGB(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.ReadRGB(), qt.Equals, [3]float32{0.15625, 1.45, 0.22222})
}

func TestReadColors(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.ReadColors(), qt.Equals, [6]float32{0.15625, 0.22222, 1.45, 0.00002, 0.15625, 0.15625})
}

func TestReadTemp(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.ReadTemp(), qt.Equals, int(25))

	fake.Registers[TempReg] = byte(0b01010011)
	c.Assert(dev.ReadTemp(), qt.Equals, int(83))

}

func defaultRegisters() []uint8 {
	return []uint8{
		HardwareVersionReg: 0x40,             // = 01000000
		TempReg:            byte(0b00011001), // = 25
		VCalReg:            byte(0b00111110),
		VCalReg + 0x01:     byte(0b00100000),
		VCalReg + 0x02:     byte(0b00000000),
		VCalReg + 0x03:     byte(0b00000000), // = 0.15625
		BCalReg:            byte(0b00111110),
		BCalReg + 0x01:     byte(0b01100011),
		BCalReg + 0x02:     byte(0b10001101),
		BCalReg + 0x03:     byte(0b10100100), // = 0.22222
		GCalReg:            byte(0b00111111),
		GCalReg + 0x01:     byte(0b10111001),
		GCalReg + 0x02:     byte(0b10011001),
		GCalReg + 0x03:     byte(0b10011010), // = 1.45
		YCalReg:            byte(0b00110111),
		YCalReg + 0x01:     byte(0b10100111),
		YCalReg + 0x02:     byte(0b11000101),
		YCalReg + 0x03:     byte(0b10101100), // = 0.00002
		OCalReg:            byte(0b00111110),
		OCalReg + 0x01:     byte(0b00100000),
		OCalReg + 0x02:     byte(0b00000000),
		OCalReg + 0x03:     byte(0b00000000), // = 0.15625
		RCalReg:            byte(0b00111110),
		RCalReg + 0x01:     byte(0b00100000),
		RCalReg + 0x02:     byte(0b00000000),
		RCalReg + 0x03:     byte(0b00000000), // = 0.15625
	}
}
