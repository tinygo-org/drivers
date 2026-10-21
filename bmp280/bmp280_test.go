package bmp280

import (
	"errors"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers"
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
	fake := newFake(c)
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.Connected(), qt.IsTrue)

	fake.Registers[REG_ID] = 0x60
	c.Assert(dev.Connected(), qt.IsFalse)
}

func TestReset(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFake(c)
	bus.AddDevice(fake)

	dev := New(bus)
	dev.Reset()
	c.Assert(fake.Registers[REG_RESET], qt.Equals, uint8(CMD_RESET))
}

func TestConfigure(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFake(c)
	bus.AddDevice(fake)

	dev := New(bus)
	dev.Configure(STANDBY_125MS, FILTER_4X, SAMPLING_2X, SAMPLING_16X, MODE_NORMAL)

	c.Assert(fake.Registers[REG_CONFIG], qt.Equals, uint8(0b010_010_00))
	c.Assert(fake.Registers[REG_CTRL_MEAS], qt.Equals, uint8(0b010_101_11))
	c.Assert(dev.cali, qt.Equals, calibrationCoefficients{
		t1: 27504, t2: 26435, t3: -1000,
		p1: 36477, p2: -10685, p3: 3024, p4: 2855, p5: 140,
		p6: -7, p7: 15500, p8: -14600, p9: 6000,
	})
}

func TestReadTemperature(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFake(c)
	bus.AddDevice(fake)

	dev := New(bus)
	dev.Configure(STANDBY_1MS, FILTER_OFF, SAMPLING_1X, SAMPLING_1X, MODE_NORMAL)

	temp, err := dev.ReadTemperature()
	c.Assert(err, qt.IsNil)
	c.Assert(temp, qt.Equals, int32(25080))
}

func TestReadPressure(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFake(c)
	bus.AddDevice(fake)

	dev := New(bus)
	dev.Configure(STANDBY_1MS, FILTER_OFF, SAMPLING_1X, SAMPLING_1X, MODE_NORMAL)

	pres, err := dev.ReadPressure()
	c.Assert(err, qt.IsNil)
	c.Assert(pres, qt.Equals, int32(100656000))
}

func TestUpdate(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFake(c)
	bus.AddDevice(fake)

	dev := New(bus)
	dev.Configure(STANDBY_1MS, FILTER_OFF, SAMPLING_1X, SAMPLING_1X, MODE_FORCED)

	c.Assert(dev.Update(drivers.Humidity), qt.IsNil)
	temp, pres := dev.Measurements()
	c.Assert(temp, qt.Equals, int32(0))
	c.Assert(pres, qt.Equals, int32(0))

	c.Assert(dev.Update(drivers.Temperature|drivers.Pressure), qt.IsNil)
	temp, pres = dev.Measurements()
	c.Assert(temp, qt.Equals, int32(25080))
	c.Assert(pres, qt.Equals, int32(100656000))

	fake.Err = errors.New("bus error")
	c.Assert(dev.Update(drivers.Temperature), qt.Equals, fake.Err)
}

func TestUpdateNormalModeIgnoresStatus(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFake(c)
	bus.AddDevice(fake)

	dev := New(bus)
	dev.Configure(STANDBY_1MS, FILTER_OFF, SAMPLING_1X, SAMPLING_1X, MODE_NORMAL)
	// A conversion is always running in normal mode with a short standby.
	fake.Registers[REG_STATUS] = statusMeasuring

	c.Assert(dev.Update(drivers.Temperature|drivers.Pressure), qt.IsNil)
	temp, pres := dev.Measurements()
	c.Assert(temp, qt.Equals, int32(25080))
	c.Assert(pres, qt.Equals, int32(100656000))
}

func TestReadTemperatureLargeT1(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFake(c)
	bus.AddDevice(fake)

	// dig_T1 = 40000 and adc_T = 700000. Expected value from the Bosch 32-bit reference code.
	fake.Registers[REG_CALI], fake.Registers[REG_CALI+1] = 0x40, 0x9C
	copy(fake.Registers[REG_TEMP:], []uint8{0xAA, 0xE6, 0x00})

	dev := New(bus)
	dev.Configure(STANDBY_1MS, FILTER_OFF, SAMPLING_1X, SAMPLING_1X, MODE_NORMAL)

	temp, err := dev.ReadTemperature()
	c.Assert(err, qt.IsNil)
	c.Assert(temp, qt.Equals, int32(18870))
}

func TestAddressLow(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice8(c, AddressLow)
	fake.Registers[REG_ID] = CHIP_ID
	bus.AddDevice(fake)

	dev := New(bus)
	dev.Address = AddressLow
	c.Assert(dev.Connected(), qt.IsTrue)
}

func TestReadAltitude(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFake(c)
	bus.AddDevice(fake)

	dev := New(bus)
	dev.Configure(STANDBY_1MS, FILTER_OFF, SAMPLING_1X, SAMPLING_1X, MODE_NORMAL)

	alt, err := dev.ReadAltitude()
	c.Assert(err, qt.IsNil)
	c.Assert(alt, qt.Equals, int32(55))
}

func TestSetMode(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFake(c)
	bus.AddDevice(fake)

	dev := New(bus)
	dev.Configure(STANDBY_1MS, FILTER_OFF, SAMPLING_4X, SAMPLING_8X, MODE_SLEEP)
	fake.Registers[REG_CALI] = 0

	err := dev.SetMode(MODE_NORMAL)
	c.Assert(err, qt.IsNil)
	c.Assert(dev.Mode, qt.Equals, MODE_NORMAL)
	c.Assert(fake.Registers[REG_CTRL_MEAS], qt.Equals, uint8(0b011_100_11))
	c.Assert(dev.cali.t1, qt.Equals, uint16(27504))
}

func TestStatusReservedBits(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFake(c)
	bus.AddDevice(fake)

	dev := New(bus)
	dev.Configure(STANDBY_1MS, FILTER_OFF, SAMPLING_1X, SAMPLING_1X, MODE_NORMAL)

	// Only bits 3 and 0 mean the data is not ready. Datasheet section 4.3.3.
	fake.Registers[REG_STATUS] = 0xF6
	temp, err := dev.ReadTemperature()
	c.Assert(err, qt.IsNil)
	c.Assert(temp, qt.Equals, int32(25080))
}

func TestMeasurementTime(t *testing.T) {
	c := qt.New(t)
	dev := New(tester.NewI2CBus(c))

	// Max values from datasheet section 3.8.1, table 13.
	dev.Temperature, dev.Pressure = SAMPLING_1X, SAMPLING_1X
	c.Assert(dev.measurementTime(), qt.Equals, 6425*time.Microsecond)
	dev.Temperature, dev.Pressure = SAMPLING_2X, SAMPLING_16X
	c.Assert(dev.measurementTime(), qt.Equals, 43225*time.Microsecond)
	dev.Temperature, dev.Pressure = SAMPLING_1X, SAMPLING_SKIPPED
	c.Assert(dev.measurementTime(), qt.Equals, 3550*time.Microsecond)
}

func TestForcedMode(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFake(c)
	bus.AddDevice(fake)

	dev := New(bus)
	dev.Configure(STANDBY_1MS, FILTER_OFF, SAMPLING_4X, SAMPLING_8X, MODE_SLEEP)
	c.Assert(fake.Registers[REG_CTRL_MEAS], qt.Equals, uint8(0b011_100_00))

	_, err := dev.ReadTemperature()
	c.Assert(err, qt.IsNil)
	c.Assert(fake.Registers[REG_CTRL_MEAS], qt.Equals, uint8(0b011_100_01))
}

func TestReadError(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFake(c)
	bus.AddDevice(fake)

	dev := New(bus)
	dev.Configure(STANDBY_1MS, FILTER_OFF, SAMPLING_1X, SAMPLING_1X, MODE_NORMAL)

	fake.Err = errors.New("bus error")
	_, err := dev.ReadTemperature()
	c.Assert(err, qt.Equals, fake.Err)
	_, err = dev.ReadPressure()
	c.Assert(err, qt.Equals, fake.Err)
}

// newFake returns a BMP280 loaded with the compensation example from
// datasheet section 3.12. The 32-bit pressure formula gives 100656 Pa.
func newFake(c *qt.C) *tester.I2CDevice8 {
	fake := tester.NewI2CDevice8(c, Address)
	fake.Registers[REG_ID] = CHIP_ID
	copy(fake.Registers[REG_CALI:], []uint8{
		0x70, 0x6B, 0x43, 0x67, 0x18, 0xFC,
		0x7D, 0x8E, 0x43, 0xD6, 0xD0, 0x0B, 0x27, 0x0B,
		0x8C, 0x00, 0xF9, 0xFF, 0x8C, 0x3C, 0xF8, 0xC6, 0x70, 0x17,
	})
	// adc_P = 415148 and adc_T = 519888, left aligned in 20 bits.
	copy(fake.Registers[REG_PRES:], []uint8{0x65, 0x5A, 0xC0, 0x7E, 0xED, 0x00})
	return fake
}
