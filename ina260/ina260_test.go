package ina260

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/tester"
)

func TestDefaultI2CAddress(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	dev := New(bus)
	c.Assert(dev.Address, qt.Equals, uint16(Address))
}

func TestConnected(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice16(c, Address)
	fake.Registers = defaultRegisters()
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.Connected(), qt.Equals, true)
}

func TestVoltage(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice16(c, Address)
	fake.Registers = defaultRegisters()
	fake.Registers[REG_BUSVOLTAGE] = 0x2570
	bus.AddDevice(fake)

	dev := New(bus)
	// Datasheet: 2570h = 11.98V = 11980mV = 11980000uV
	c.Assert(dev.Voltage(), qt.Equals, int32(11980000))
}

func TestCurrent(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice16(c, Address)
	fake.Registers = defaultRegisters()
	fake.Registers[REG_CURRENT] = 0x2710
	bus.AddDevice(fake)

	dev := New(bus)
	// Datasheet: 2710h = 12.5A = 12500mA = 12500000uA
	c.Assert(dev.Current(), qt.Equals, int32(12500000))
}

func TestPower(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice16(c, Address)
	fake.Registers = defaultRegisters()
	fake.Registers[REG_POWER] = 0x3A7F
	bus.AddDevice(fake)

	dev := New(bus)
	// 3A7Fh = 149.75W = 149750mW = 149750000uW
	c.Assert(dev.Power(), qt.Equals, int32(149750000))
}

// defaultRegisters returns the default values for all of the device's registers.
// set TI INA260 datasheet for power-on defaults
func defaultRegisters() map[uint8]uint16 {
	return map[uint8]uint16{
		REG_CONFIG:     0x6127,
		REG_CURRENT:    0x0000,
		REG_BUSVOLTAGE: 0x0000,
		REG_POWER:      0x0000,
		REG_MASKENABLE: 0x0000,
		REG_ALERTLIMIT: 0x0000,
		REG_MANF_ID:    0x5449,
		REG_DIE_ID:     0x2270,
	}
}
