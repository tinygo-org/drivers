package ina3221

import (
	"errors"
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/tester"
)

func defaultRegisters() map[uint8]uint16 {
	return map[uint8]uint16{
		RegConfig:         0x7127,
		RegShuntVoltage1:  0,
		RegBusVoltage1:    0,
		RegShuntVoltage2:  0,
		RegBusVoltage2:    0,
		RegShuntVoltage3:  0,
		RegBusVoltage3:    0,
		RegMaskEnable:     0x0002,
		RegManufacturerID: ManufacturerID,
		RegDieID:          DieID,
	}
}

func newDevice(c *qt.C) (*Device, *tester.I2CDevice16) {
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice16(c, Address)
	fake.Registers = defaultRegisters()
	bus.AddDevice(fake)
	dev := New(bus)
	return &dev, fake
}

func TestDefaultAddress(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	dev := New(bus)
	c.Assert(dev.Address, qt.Equals, uint16(Address))
}

func TestDefaultConfig(t *testing.T) {
	c := qt.New(t)
	c.Assert(DefaultConfig.RegisterValue(), qt.Equals, uint16(0x7127))
	c.Assert(NewConfig(0x7127), qt.Equals, DefaultConfig)
}

func TestConfigRoundTrip(t *testing.T) {
	c := qt.New(t)
	cfg := Config{
		Channels:      EnableChannel1 | EnableChannel3,
		Average:       Average64,
		BusConvTime:   ConvTime8244us,
		ShuntConvTime: ConvTime140us,
		Mode:          ModeSingleShuntBus,
	}
	// 0b0101_011_111_000_011
	c.Assert(cfg.RegisterValue(), qt.Equals, uint16(0x57C3))
	c.Assert(NewConfig(cfg.RegisterValue()), qt.Equals, cfg)
}

func TestConnected(t *testing.T) {
	c := qt.New(t)
	dev, fake := newDevice(c)
	c.Assert(dev.Connected(), qt.IsTrue)

	fake.Registers[RegDieID] = 0x2270
	c.Assert(dev.Connected(), qt.IsFalse)

	fake.Registers[RegDieID] = DieID
	fake.Registers[RegManufacturerID] = 0
	c.Assert(dev.Connected(), qt.IsFalse)
}

func TestConfigure(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		c := qt.New(t)
		dev, fake := newDevice(c)
		cfg := DefaultConfig
		cfg.Average = Average16
		cfg.Mode = ModeContBus
		c.Assert(dev.Configure(cfg), qt.IsNil)
		c.Assert(fake.Registers[RegConfig], qt.Equals, cfg.RegisterValue())
		c.Assert(dev.Config(), qt.Equals, cfg)
	})

	t.Run("not connected", func(t *testing.T) {
		c := qt.New(t)
		dev, fake := newDevice(c)
		fake.Registers[RegManufacturerID] = 0
		c.Assert(dev.Configure(DefaultConfig), qt.Equals, ErrDeviceNotConnected)
	})

	t.Run("bus error", func(t *testing.T) {
		c := qt.New(t)
		dev, fake := newDevice(c)
		fake.Err = errors.New("bus error")
		c.Assert(dev.Configure(DefaultConfig), qt.Equals, ErrDeviceNotConnected)
	})
}

func TestReset(t *testing.T) {
	c := qt.New(t)
	dev, fake := newDevice(c)
	cfg := DefaultConfig
	cfg.Mode = ModePowerDown
	c.Assert(dev.Configure(cfg), qt.IsNil)

	c.Assert(dev.Reset(), qt.IsNil)
	c.Assert(fake.Registers[RegConfig], qt.Equals, uint16(0x8000))
	c.Assert(dev.Config(), qt.Equals, DefaultConfig)
}

func TestTrigger(t *testing.T) {
	t.Run("single shot", func(t *testing.T) {
		c := qt.New(t)
		dev, fake := newDevice(c)
		cfg := DefaultConfig
		cfg.Mode = ModeSingleShuntBus
		c.Assert(dev.Configure(cfg), qt.IsNil)

		fake.Registers[RegConfig] = 0
		c.Assert(dev.Trigger(), qt.IsNil)
		c.Assert(fake.Registers[RegConfig], qt.Equals, cfg.RegisterValue())
	})

	t.Run("continuous", func(t *testing.T) {
		c := qt.New(t)
		dev, fake := newDevice(c)
		c.Assert(dev.Configure(DefaultConfig), qt.IsNil)

		fake.Registers[RegConfig] = 0
		c.Assert(dev.Trigger(), qt.IsNil)
		c.Assert(fake.Registers[RegConfig], qt.Equals, uint16(0))
	})
}

func TestModeSingleShot(t *testing.T) {
	c := qt.New(t)
	for mode, want := range map[Mode]bool{
		ModePowerDown:      false,
		ModeSingleShunt:    true,
		ModeSingleBus:      true,
		ModeSingleShuntBus: true,
		4:                  false,
		ModeContShunt:      false,
		ModeContBus:        false,
		ModeContShuntBus:   false,
	} {
		c.Assert(mode.SingleShot(), qt.Equals, want, qt.Commentf("mode %d", mode))
	}
}

func TestConversionReady(t *testing.T) {
	c := qt.New(t)
	dev, fake := newDevice(c)

	ready, err := dev.ConversionReady()
	c.Assert(err, qt.IsNil)
	c.Assert(ready, qt.IsFalse)

	fake.Registers[RegMaskEnable] = 0x0003
	ready, err = dev.ConversionReady()
	c.Assert(err, qt.IsNil)
	c.Assert(ready, qt.IsTrue)
}

func TestBusVoltage(t *testing.T) {
	for _, tc := range []struct {
		name    string
		channel int
		reg     uint8
		val     uint16
		want    int32
	}{
		{"ch1 max code", Channel1, RegBusVoltage1, 0x7FF8, 32760000},
		{"ch2 12V", Channel2, RegBusVoltage2, 1500 << 3, 12000000},
		{"ch3 3.3V", Channel3, RegBusVoltage3, 412 << 3, 3296000},
		{"ch1 negative", Channel1, RegBusVoltage1, 0xFFF8, -8000},
		{"ch2 ignores low bits", Channel2, RegBusVoltage2, (1 << 3) | 0x7, 8000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			dev, fake := newDevice(c)
			fake.Registers[tc.reg] = tc.val
			v, err := dev.BusVoltage(tc.channel)
			c.Assert(err, qt.IsNil)
			c.Assert(v, qt.Equals, tc.want)
		})
	}
}

func TestShuntVoltage(t *testing.T) {
	for _, tc := range []struct {
		name    string
		channel int
		reg     uint8
		val     uint16
		want    int32
	}{
		{"ch1 full scale", Channel1, RegShuntVoltage1, 0x7FF8, 163800},
		{"ch2 negative full scale", Channel2, RegShuntVoltage2, 0x8000, -163840},
		{"ch3 one LSB", Channel3, RegShuntVoltage3, 1 << 3, 40},
		{"ch1 minus one LSB", Channel1, RegShuntVoltage1, 0xFFF8, -40},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			dev, fake := newDevice(c)
			fake.Registers[tc.reg] = tc.val
			v, err := dev.ShuntVoltage(tc.channel)
			c.Assert(err, qt.IsNil)
			c.Assert(v, qt.Equals, tc.want)
		})
	}
}

func TestCurrent(t *testing.T) {
	c := qt.New(t)
	dev, fake := newDevice(c)

	// 10mV across the default 100 milliohm shunt is 100mA.
	fake.Registers[RegShuntVoltage1] = 250 << 3
	i, err := dev.Current(Channel1)
	c.Assert(err, qt.IsNil)
	c.Assert(i, qt.Equals, int32(100000))

	// 10mV across 10 milliohm is 1A.
	c.Assert(dev.SetShuntResistance(Channel1, 10), qt.IsNil)
	i, err = dev.Current(Channel1)
	c.Assert(err, qt.IsNil)
	c.Assert(i, qt.Equals, int32(1000000))

	fake.Registers[RegShuntVoltage2] = 0xFFF8
	i, err = dev.Current(Channel2)
	c.Assert(err, qt.IsNil)
	c.Assert(i, qt.Equals, int32(-400))
}

func TestPower(t *testing.T) {
	c := qt.New(t)
	dev, fake := newDevice(c)

	// 12V at 100mA is 1.2W.
	fake.Registers[RegBusVoltage3] = 1500 << 3
	fake.Registers[RegShuntVoltage3] = 250 << 3
	p, err := dev.Power(Channel3)
	c.Assert(err, qt.IsNil)
	c.Assert(p, qt.Equals, int32(1200000))
}

func TestMeasurements(t *testing.T) {
	c := qt.New(t)
	dev, fake := newDevice(c)
	fake.Registers[RegBusVoltage2] = 625 << 3
	fake.Registers[RegShuntVoltage2] = 500 << 3
	c.Assert(dev.SetShuntResistance(Channel2, 50), qt.IsNil)

	bus, shunt, current, power, err := dev.Measurements(Channel2)
	c.Assert(err, qt.IsNil)
	c.Assert(bus, qt.Equals, int32(5000000))
	c.Assert(shunt, qt.Equals, int32(20000))
	c.Assert(current, qt.Equals, int32(400000))
	c.Assert(power, qt.Equals, int32(2000000))
}

func TestInvalidChannel(t *testing.T) {
	c := qt.New(t)
	dev, _ := newDevice(c)
	for _, ch := range []int{0, 4, -1} {
		_, err := dev.BusVoltage(ch)
		c.Assert(err, qt.Equals, ErrInvalidChannel)
		_, err = dev.ShuntVoltage(ch)
		c.Assert(err, qt.Equals, ErrInvalidChannel)
		_, err = dev.Current(ch)
		c.Assert(err, qt.Equals, ErrInvalidChannel)
		_, err = dev.Power(ch)
		c.Assert(err, qt.Equals, ErrInvalidChannel)
		_, _, _, _, err = dev.Measurements(ch)
		c.Assert(err, qt.Equals, ErrInvalidChannel)
		c.Assert(dev.SetShuntResistance(ch, 100), qt.Equals, ErrInvalidChannel)
	}
}

func TestSetShuntResistanceInvalid(t *testing.T) {
	c := qt.New(t)
	dev, _ := newDevice(c)
	c.Assert(dev.SetShuntResistance(Channel1, 0), qt.Equals, ErrInvalidResistance)
	c.Assert(dev.SetShuntResistance(Channel1, -5), qt.Equals, ErrInvalidResistance)
}

func TestBusError(t *testing.T) {
	c := qt.New(t)
	dev, fake := newDevice(c)
	busErr := errors.New("bus error")
	fake.Err = busErr

	_, err := dev.BusVoltage(Channel1)
	c.Assert(err, qt.Equals, busErr)
	_, err = dev.ShuntVoltage(Channel1)
	c.Assert(err, qt.Equals, busErr)
	_, err = dev.ConversionReady()
	c.Assert(err, qt.Equals, busErr)
	c.Assert(dev.Reset(), qt.Equals, busErr)
	c.Assert(dev.Connected(), qt.IsFalse)
}

func TestReadDoesNotAllocate(t *testing.T) {
	c := qt.New(t)
	dev, _ := newDevice(c)
	allocs := testing.AllocsPerRun(100, func() {
		dev.BusVoltage(2)
		dev.Current(2)
	})
	c.Assert(allocs, qt.Equals, float64(0))
}
