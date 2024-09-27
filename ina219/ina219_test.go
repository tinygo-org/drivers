package ina219

import (
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/tester"
)

func TestDefaultAddress(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	dev := New(bus)
	c.Assert(dev.Address, qt.Equals, uint16(Address))
}

func TestBusVoltage(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		c := qt.New(t)
		bus := tester.NewI2CBus(c)
		fake := tester.NewI2CDevice16(c, Address)
		fake.Registers = map[uint8]uint16{
			RegBusVoltage: (4200 << 3) / 4, // 4.2V
		}
		bus.AddDevice(fake)

		dev := New(bus)
		voltage, err := dev.BusVoltage()
		c.Assert(err, qt.IsNil)
		c.Assert(voltage, qt.Equals, int16(4200))
	})

	t.Run("overflow", func(t *testing.T) {
		c := qt.New(t)
		bus := tester.NewI2CBus(c)
		fake := tester.NewI2CDevice16(c, Address)
		fake.Registers = map[uint8]uint16{
			RegBusVoltage: (1 >> 0), // overflow
		}
		bus.AddDevice(fake)

		dev := New(bus)
		_, err := dev.BusVoltage()
		c.Assert(err, qt.Not(qt.IsNil))
		c.Assert(err, qt.ErrorMatches, ErrOverflow{}.Error())
	})

	t.Run("not ready", func(t *testing.T) {
		c := qt.New(t)
		bus := tester.NewI2CBus(c)
		fake := tester.NewI2CDevice16(c, Address)
		fake.Registers = map[uint8]uint16{
			RegBusVoltage: ((4200 << 3) / 4) | (1 << 1), // not ready
		}
		bus.AddDevice(fake)

		dev := New(bus)
		dev.config.Mode = ModeTrigBus
		_, err := dev.BusVoltage()
		c.Assert(err, qt.Not(qt.IsNil))
		c.Assert(err, qt.ErrorMatches, ErrNotReady{}.Error())
	})
}

func TestShuntVoltage(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice16(c, Address)
	fake.Registers = map[uint8]uint16{
		RegShuntVoltage: 0x1234,
	}
	bus.AddDevice(fake)

	dev := New(bus)
	voltage, err := dev.ShuntVoltage()
	c.Assert(err, qt.IsNil)
	c.Assert(voltage, qt.Equals, int16(0x1234))
}

func TestCurrent(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice16(c, Address)
	fake.Registers = map[uint8]uint16{
		RegCurrent: 420 * 6.9, // 420mA
	}
	bus.AddDevice(fake)

	dev := New(bus)
	dev.config.CurrentDivider = 6.9
	current, err := dev.Current()
	c.Assert(err, qt.IsNil)
	c.Assert(current, qt.Equals, float32(420))
}

func TestPower(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice16(c, Address)
	fake.Registers = map[uint8]uint16{
		RegPower: 420 / 0.8, // 420mW
	}
	bus.AddDevice(fake)

	dev := New(bus)
	dev.config.PowerMultiplier = 0.8
	power, err := dev.Power()
	c.Assert(err, qt.IsNil)
	c.Assert(power, qt.Equals, float32(420))
}

func TestReadConfig(t *testing.T) {
	// use the default configurations
	for _, tc := range []Config{
		Config16V400mA,
		Config32V2A,
		Config32V1A,
	} {
		n := fmt.Sprintf("%x/%x", tc.RegisterValue(), tc.Calibration.RegisterValue())
		t.Run(n, func(t *testing.T) {
			c := qt.New(t)
			bus := tester.NewI2CBus(c)
			fake := tester.NewI2CDevice16(c, Address)
			fake.Registers = map[uint8]uint16{
				RegConfig:      tc.RegisterValue(),
				RegCalibration: tc.Calibration.RegisterValue(),
			}
			bus.AddDevice(fake)

			dev := New(bus)
			config, err := dev.ReadConfig()
			c.Assert(err, qt.IsNil)
			c.Assert(config.BusADC, qt.Equals, tc.BusADC)
			c.Assert(config.BusVoltageRange, qt.Equals, tc.BusVoltageRange)
			c.Assert(config.Calibration, qt.Equals, tc.Calibration)
			c.Assert(config.Mode, qt.Equals, tc.Mode)
			c.Assert(config.PGA, qt.Equals, tc.PGA)
			c.Assert(config.ShuntADC, qt.Equals, tc.ShuntADC)
		})
	}
}

func TestWriteConfig(t *testing.T) {
	// use the default configurations
	for _, tc := range []Config{
		Config16V400mA,
		Config32V2A,
		Config32V1A,
	} {
		n := fmt.Sprintf("%x/%x", tc.RegisterValue(), tc.Calibration.RegisterValue())
		t.Run(n, func(t *testing.T) {
			c := qt.New(t)
			bus := tester.NewI2CBus(c)
			fake := tester.NewI2CDevice16(c, Address)
			bus.AddDevice(fake)
			fake.Registers = map[uint8]uint16{
				RegConfig:      0,
				RegCalibration: 0,
			}

			dev := New(bus)
			dev.config = tc
			err := dev.Configure()
			c.Assert(err, qt.IsNil)
			c.Assert(fake.Registers[RegConfig], qt.Equals, tc.RegisterValue())
			c.Assert(fake.Registers[RegCalibration], qt.Equals, tc.Calibration.RegisterValue())
		})
	}
}

func TestSetConfig(t *testing.T) {
	for _, tc := range []Config{
		Config16V400mA,
		Config32V2A,
		Config32V1A,
	} {
		n := fmt.Sprintf("%x/%x", tc.RegisterValue(), tc.Calibration.RegisterValue())
		t.Run(n, func(t *testing.T) {
			c := qt.New(t)
			bus := tester.NewI2CBus(c)
			dev := New(bus)
			dev.SetConfig(tc)
			c.Assert(dev.config, qt.Equals, tc)
		})
	}
}

func TestTrigger(t *testing.T) {
	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice16(c, Address)
	bus.AddDevice(fake)
	fake.Registers = map[uint8]uint16{
		RegConfig: Config32V2A.RegisterValue(),
	}

	dev := New(bus)
	dev.config = Config32V2A
	dev.config.Mode = ModeTrigBus
	err := dev.Trigger()
	c.Assert(err, qt.IsNil)
	c.Assert(fake.Registers[RegConfig], qt.Equals, dev.config.RegisterValue())
}

func TestMeasurements(t *testing.T) {
	bvVal := int16(4200)
	svVal := int16(1234)
	iVal := float32(420)
	pVal := float32(420)

	c := qt.New(t)
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice16(c, Address)
	bus.AddDevice(fake)
	fake.Registers = map[uint8]uint16{
		RegBusVoltage:   uint16(((4200 << 3) / 4) | (1 << 1)),
		RegShuntVoltage: uint16(svVal),
		RegCurrent:      uint16(iVal * Config16V400mA.CurrentDivider),
		RegPower:        uint16(pVal / Config16V400mA.PowerMultiplier),
	}

	dev := New(bus)
	dev.config = Config16V400mA
	bv, sv, i, p, err := dev.Measurements()
	c.Assert(err, qt.IsNil)
	c.Assert(bv, qt.Equals, bvVal)
	c.Assert(sv, qt.Equals, svVal)
	c.Assert(i, qt.Equals, iVal)
	c.Assert(p, qt.Equals, pVal)

}
