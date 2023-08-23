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

func TestGetViolet(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.getViolet(), qt.Equals, float32(0.15625))

	fake.Registers[VCalReg+1] = byte(0b00111111)
	c.Assert(dev.getViolet(), qt.Not(qt.Equals), float32(0.15625))

}

func TestGetBlue(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.getBlue(), qt.Equals, float32(0.22222))

	fake.Registers[BCalReg+1] = byte(0b00111111)
	c.Assert(dev.getBlue(), qt.Not(qt.Equals), float32(0.22222))

}

func TestGetGreen(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.getGreen(), qt.Equals, float32(1.45))

	fake.Registers[GCalReg+1] = byte(0b00000001)
	c.Assert(dev.getGreen(), qt.Not(qt.Equals), float32(1.45))

}

func TestGetYellow(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.getYellow(), qt.Equals, float32(0.00002))

	fake.Registers[YCalReg+1] = byte(0b00000001)
	c.Assert(dev.getYellow(), qt.Not(qt.Equals), float32(0.00002))

}

func TestGetOrange(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.getOrange(), qt.Equals, float32(0.15625))

	fake.Registers[OCalReg+1] = byte(0b00000001)
	c.Assert(dev.getOrange(), qt.Not(qt.Equals), float32(0.15625))

}

func TestGetRed(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.getRed(), qt.Equals, float32(0.15625))

	fake.Registers[RCalReg+1] = byte(0b00000001)
	c.Assert(dev.getRed(), qt.Not(qt.Equals), float32(0.15625))

}

func TestGetRGB(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.getRGB(), qt.Equals, [3]float32{0.15625, 1.45, 0.22222})
}

func TestGetColors(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.getColors(), qt.Equals, [6]float32{0.15625, 0.22222, 1.45, 0.00002, 0.15625, 0.15625})
}

func TestGetTemp(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice(c, DefaultAddress)
	copy(fake.Registers[:], defaultRegisters())
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.getTemp(), qt.Equals, int(25))

	fake.Registers[TempReg] = byte(0b01010011)
	c.Assert(dev.getTemp(), qt.Equals, int(83))

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
