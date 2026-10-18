package bmm150

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/tester"
)

func TestDefaultAddress(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	dev := New(bus)
	c.Assert(dev.Address, qt.Equals, uint16(DefaultAddress))
}

func TestConnected(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFakeDevice(c)
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.Connected(), qt.Equals, true)
	c.Assert(fake.Registers[regPowerCtrl], qt.Equals, uint8(powerOn))

	fake.Registers[regChipID] = 0x99
	c.Assert(dev.Connected(), qt.Equals, false)
}

func TestConfigure(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFakeDevice(c)
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.Configure(Config{}), qt.IsNil)
	c.Assert(fake.Registers[regRepXY], qt.Equals, uint8(0x04))
	c.Assert(fake.Registers[regRepZ], qt.Equals, uint8(0x0E))
	c.Assert(fake.Registers[regOpMode], qt.Equals, uint8(0x00))

	c.Assert(dev.trim, qt.Equals, trimData{
		x1: 1, y1: -2, x2: 26, y2: 24,
		xy1: 29, xy2: -3,
		z1: 24747, z2: 763, z3: -12, z4: 5,
		xyz1: 6539,
	})

	c.Assert(dev.Configure(Config{DataRate: DataRate20Hz, Preset: HighAccuracy}), qt.IsNil)
	c.Assert(fake.Registers[regRepXY], qt.Equals, uint8(0x17))
	c.Assert(fake.Registers[regRepZ], qt.Equals, uint8(0x52))
	c.Assert(fake.Registers[regOpMode], qt.Equals, uint8(0x28))
}

func TestConfigureNotConnected(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFakeDevice(c)
	fake.Registers[regChipID] = 0
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.Configure(Config{}), qt.Equals, errNotConnected)
}

func TestMagneticField(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFakeDevice(c)
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.Configure(Config{}), qt.IsNil)

	// Expected values are computed with the integer formulas of the Bosch
	// BMM150 Sensor API without the final division by 16.
	tests := []struct {
		x, y, z, rhall int16
		wx, wy, wz     int32
	}{
		{200, -150, 500, 6500, 73187, -54937, 178687},
		{-3000, 2500, -8000, 6400, -1091937, 899625, -2047937},
	}
	for _, tc := range tests {
		setData(fake, tc.x, tc.y, tc.z, tc.rhall)
		c.Assert(dev.Update(drivers.MagneticField), qt.IsNil)
		x, y, z := dev.MagneticField()
		c.Assert(x, qt.Equals, tc.wx)
		c.Assert(y, qt.Equals, tc.wy)
		c.Assert(z, qt.Equals, tc.wz)
	}
}

func TestOverflow(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := newFakeDevice(c)
	bus.AddDevice(fake)

	dev := New(bus)
	c.Assert(dev.Configure(Config{}), qt.IsNil)

	setData(fake, overflowXY, 0, 0, 6500)
	c.Assert(dev.Update(drivers.MagneticField), qt.Equals, errOverflow)

	setData(fake, 0, 0, overflowZ, 6500)
	c.Assert(dev.Update(drivers.MagneticField), qt.Equals, errOverflow)
}

func newFakeDevice(c *qt.C) *tester.I2CDevice8 {
	fake := tester.NewI2CDevice(c, DefaultAddress)
	fake.Registers[regChipID] = chipID
	copy(fake.Registers[regTrim:], []uint8{
		0x01, 0xFE, // DIG_X1, DIG_Y1
		0, 0, 0, // reserved
		0x05, 0x00, // DIG_Z4
		0x1A, 0x18, // DIG_X2, DIG_Y2
		0, 0, // reserved
		0xFB, 0x02, // DIG_Z2
		0xAB, 0x60, // DIG_Z1
		0x8B, 0x99, // DIG_XYZ1 with bit 15 set
		0xF4, 0xFF, // DIG_Z3
		0xFD, 0x1D, // DIG_XY2, DIG_XY1
	})
	return fake
}

func setData(fake *tester.I2CDevice8, x, y, z, rhall int16) {
	put := func(reg uint8, v uint16) {
		fake.Registers[reg] = uint8(v)
		fake.Registers[reg+1] = uint8(v >> 8)
	}
	put(regDataX, uint16(x)<<3)
	put(regDataX+2, uint16(y)<<3)
	put(regDataX+4, uint16(z)<<1)
	put(regDataX+6, uint16(rhall)<<2|1)
}
