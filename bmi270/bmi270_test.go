package bmi270

import (
	"errors"
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers"
)

// fake emulates the BMI270 register behavior used by the driver.
type fake struct {
	c        *qt.C
	regs     [128]byte
	config   [8192]byte
	written  int
	resets   int
	initDone bool
	csLow    bool
	spiMode  bool
	writes   []byte // registers written outside of the config upload
	corrupt  bool
	err      error
	ops      int
	failAt   int // fail the Nth bus operation when non-zero
}

var errBus = errors.New("bus error")

func (f *fake) fail() error {
	f.ops++
	if f.failAt != 0 && f.ops == f.failAt {
		return errBus
	}
	return f.err
}

func newFake(c *qt.C) *fake {
	f := &fake{c: c}
	f.reset()
	return f
}

func (f *fake) reset() {
	f.regs = [128]byte{}
	f.regs[RegChipID] = ChipID
	f.regs[RegTemperature+1] = 0x80
	f.regs[RegPwrConf] = 0x03
	f.regs[RegAccelConf] = 0xA8
	f.regs[RegAccelRange] = 0x02
	f.regs[RegGyroConf] = 0xA9
	f.initDone = false
	f.spiMode = false
	f.written = 0
}

func (f *fake) read(reg uint8, data []byte) {
	for i := range data {
		data[i] = f.regs[int(reg)+i]
	}
}

func (f *fake) write(reg uint8, data []byte) {
	switch reg {
	case RegInitData:
		f.c.Assert(f.regs[RegPwrConf]&pwrConfAdvPowerSave, qt.Equals, uint8(0), qt.Commentf("burst write with power save on"))
		addr := (int(f.regs[RegInitAddr0]) | int(f.regs[RegInitAddr1])<<4) * 2
		copy(f.config[addr:], data)
		if f.corrupt {
			f.config[addr] ^= 0xFF
		}
		f.written += len(data)
		return
	case RegInitAddr0:
		f.regs[RegInitAddr0] = data[0] & 0x0F
		if len(data) > 1 {
			f.regs[RegInitAddr1] = data[1]
		}
		return
	}
	f.c.Assert(len(data), qt.Equals, 1)
	f.writes = append(f.writes, reg)
	switch reg {
	case RegCmd:
		if data[0] == cmdSoftReset {
			f.resets++
			f.reset()
		}
		return
	case RegInitCtrl:
		f.regs[reg] = data[0]
		if data[0] == 1 {
			f.c.Assert(f.initDone, qt.IsFalse, qt.Commentf("INIT_CTRL set twice"))
			f.initDone = true
			if f.written == len(f.config) && string(f.config[:]) == configFile {
				f.regs[RegInternalStatus] = internalStatusInitOK
			} else {
				f.regs[RegInternalStatus] = 0x02
			}
		}
		return
	}
	f.regs[reg] = data[0]
}

// Tx implements drivers.I2C.
func (f *fake) Tx(addr uint16, w, r []byte) error {
	f.c.Assert(addr, qt.Equals, uint16(Address))
	if err := f.fail(); err != nil {
		return err
	}
	if len(r) > 0 {
		f.c.Assert(len(w), qt.Equals, 1)
		f.read(w[0], r)
		return nil
	}
	f.write(w[0], w[1:])
	return nil
}

type fakeSPI struct{ *fake }

// Tx implements drivers.SPI.
func (s fakeSPI) Tx(w, r []byte) error {
	f := s.fake
	f.c.Assert(f.csLow, qt.IsTrue, qt.Commentf("cs not low"))
	if err := f.fail(); err != nil {
		return err
	}
	if r != nil {
		f.c.Assert(len(r), qt.Equals, len(w))
	}
	reg := w[0] &^ spiRead
	if w[0]&spiRead == 0 {
		f.c.Assert(r, qt.IsNil)
		f.write(reg, w[1:])
		return nil
	}
	if !f.spiMode {
		// The first read only selects SPI and returns garbage.
		f.spiMode = true
		for i := range r {
			r[i] = 0xFF
		}
		return nil
	}
	r[0], r[1] = 0xFF, 0xFF
	f.read(reg, r[2:])
	return nil
}

func (s fakeSPI) Transfer(b byte) (byte, error) {
	s.c.Fatalf("unexpected Transfer")
	return 0, nil
}

func (s fakeSPI) Set(level bool) {
	s.csLow = !level
}

func newI2C(c *qt.C) (*Device, *fake) {
	f := newFake(c)
	dev := NewI2C(f)
	return &dev, f
}

func newSPI(c *qt.C) (*Device, *fake) {
	f := newFake(c)
	s := fakeSPI{f}
	dev := NewSPI(s, s)
	return &dev, f
}

func forEachBus(t *testing.T, fn func(c *qt.C, dev *Device, f *fake)) {
	t.Run("i2c", func(t *testing.T) {
		c := qt.New(t)
		dev, f := newI2C(c)
		fn(c, dev, f)
	})
	t.Run("spi", func(t *testing.T) {
		c := qt.New(t)
		dev, f := newSPI(c)
		fn(c, dev, f)
		c.Assert(f.csLow, qt.IsFalse)
	})
}

func TestConfigFile(t *testing.T) {
	c := qt.New(t)
	c.Assert(len(configFile), qt.Equals, 8192)
	c.Assert(len(configFile)%burstLen, qt.Equals, 0)
}

func TestNew(t *testing.T) {
	c := qt.New(t)
	dev, _ := newI2C(c)
	c.Assert(dev.Address, qt.Equals, uint16(Address))
	dev, _ = newSPI(c)
	c.Assert(dev.spi, qt.Not(qt.IsNil))
}

func TestConnected(t *testing.T) {
	forEachBus(t, func(c *qt.C, dev *Device, f *fake) {
		c.Assert(dev.Connected(), qt.IsTrue)
		f.regs[RegChipID] = 0x26
		c.Assert(dev.Connected(), qt.IsFalse)
		f.regs[RegChipID] = ChipID
		f.err = errors.New("bus error")
		c.Assert(dev.Connected(), qt.IsFalse)
	})
}

func TestConfigure(t *testing.T) {
	forEachBus(t, func(c *qt.C, dev *Device, f *fake) {
		c.Assert(dev.Configure(Config{}), qt.IsNil)
		c.Assert(f.resets, qt.Equals, 1)
		c.Assert(f.written, qt.Equals, len(configFile))
		c.Assert(string(f.config[:]), qt.Equals, configFile)
		c.Assert(f.regs[RegInitCtrl], qt.Equals, uint8(1))
		c.Assert(f.regs[RegPwrCtrl], qt.Equals, uint8(0x0E))
		c.Assert(f.regs[RegAccelConf], qt.Equals, uint8(0xA8))
		c.Assert(f.regs[RegAccelRange], qt.Equals, uint8(0x02))
		c.Assert(f.regs[RegGyroConf], qt.Equals, uint8(0xE9))
		c.Assert(f.regs[RegGyroRange], qt.Equals, uint8(0x00))
		c.Assert(f.regs[RegPwrConf], qt.Equals, uint8(0x02))
		c.Assert(f.writes, qt.DeepEquals, []byte{
			RegCmd, RegPwrConf, RegInitCtrl, RegInitCtrl,
			RegPwrCtrl, RegAccelConf, RegAccelRange, RegGyroConf, RegGyroRange, RegPwrConf,
		})

		// A second Configure must reset again before loading the config.
		c.Assert(dev.Configure(Config{AccelRange: AccelRange2G, GyroRange: GyroRange125}), qt.IsNil)
		c.Assert(f.resets, qt.Equals, 2)
		c.Assert(f.regs[RegAccelRange], qt.Equals, uint8(0x00))
		c.Assert(f.regs[RegGyroRange], qt.Equals, uint8(0x04))
	})
}

func TestConfigureOptions(t *testing.T) {
	c := qt.New(t)
	dev, f := newI2C(c)
	cfg := Config{
		AccelRange: AccelRange16G,
		AccelRate:  AccelRate1600Hz,
		GyroRange:  GyroRange500,
		GyroRate:   GyroRate3200Hz,
	}
	c.Assert(dev.Configure(cfg), qt.IsNil)
	c.Assert(f.regs[RegAccelConf], qt.Equals, uint8(0xAC))
	c.Assert(f.regs[RegAccelRange], qt.Equals, uint8(0x03))
	c.Assert(f.regs[RegGyroConf], qt.Equals, uint8(0xED))
	c.Assert(f.regs[RegGyroRange], qt.Equals, uint8(0x02))
}

func TestConfigureInvalid(t *testing.T) {
	for _, cfg := range []Config{
		{AccelRange: AccelRange16G + 1},
		{AccelRate: AccelRate12_5Hz - 1},
		{AccelRate: AccelRate1600Hz + 1},
		{GyroRange: GyroRange125 + 1},
		{GyroRate: GyroRate25Hz - 1},
		{GyroRate: GyroRate3200Hz + 1},
	} {
		c := qt.New(t)
		dev, f := newI2C(c)
		c.Assert(dev.Configure(cfg), qt.Equals, ErrInvalidConfig)
		c.Assert(f.writes, qt.HasLen, 0)
	}
}

func TestConfigureErrors(t *testing.T) {
	forEachBus(t, func(c *qt.C, dev *Device, f *fake) {
		f.regs[RegChipID] = 0
		c.Assert(dev.Configure(Config{}), qt.Equals, ErrDeviceNotConnected)
	})

	t.Run("init failed", func(t *testing.T) {
		c := qt.New(t)
		dev, f := newI2C(c)
		f.corrupt = true
		c.Assert(dev.Configure(Config{}), qt.Equals, ErrInitFailed)
	})

	t.Run("bus error", func(t *testing.T) {
		c := qt.New(t)
		dev, f := newI2C(c)
		c.Assert(dev.Connected(), qt.IsTrue)
		f.err = errors.New("bus error")
		c.Assert(dev.uploadConfig(), qt.Equals, f.err)
		c.Assert(dev.waitInit(), qt.Equals, f.err)
		c.Assert(dev.writeRegister(RegPwrConf, 0), qt.Equals, f.err)
	})
}

func TestConfigureBusErrorAtEachStep(t *testing.T) {
	forEachBus(t, func(c *qt.C, dev *Device, f *fake) {
		for n := 1; ; n++ {
			f.reset()
			f.ops = 0
			f.failAt = n
			err := dev.Configure(Config{})
			if f.ops < n {
				c.Assert(err, qt.IsNil)
				break
			}
			// The SPI interface selection reads ignore errors.
			if err != nil && err != errBus && err != ErrDeviceNotConnected {
				c.Fatalf("step %d: unexpected error %v", n, err)
			}
		}
	})
}

func TestReset(t *testing.T) {
	forEachBus(t, func(c *qt.C, dev *Device, f *fake) {
		c.Assert(dev.Configure(Config{AccelRange: AccelRange2G, GyroRange: GyroRange125}), qt.IsNil)
		c.Assert(dev.Reset(), qt.IsNil)
		c.Assert(f.resets, qt.Equals, 2)
		f.err = errBus
		c.Assert(dev.Reset(), qt.Equals, errBus)
		c.Assert(dev.accelRange, qt.Equals, uint8(AccelRange8G-1))
		c.Assert(dev.gyroRange, qt.Equals, uint8(GyroRange2000-1))
	})
}

func putInt16(b []byte, v int16) {
	b[0] = byte(v)
	b[1] = byte(uint16(v) >> 8)
}

func TestAcceleration(t *testing.T) {
	for _, tc := range []struct {
		r   AccelRange
		raw int16
		ug  int32
	}{
		{AccelRange2G, 16384, 1000000},
		{AccelRange2G, -16384, -1000000},
		{AccelRange4G, 8192, 1000000},
		{AccelRange8G, 4096, 1000000},
		{AccelRange16G, 2048, 1000000},
		{AccelRange16G, -32768, -16000000},
		{AccelRange16G, 32767, 15999511},
	} {
		forEachBus(t, func(c *qt.C, dev *Device, f *fake) {
			c.Assert(dev.Configure(Config{AccelRange: tc.r}), qt.IsNil)
			putInt16(f.regs[RegAccelData:], tc.raw)
			putInt16(f.regs[RegAccelData+2:], tc.raw)
			putInt16(f.regs[RegAccelData+4:], 0)
			x, y, z, err := dev.ReadAcceleration()
			c.Assert(err, qt.IsNil)
			c.Assert(x, qt.Equals, tc.ug)
			c.Assert(y, qt.Equals, tc.ug)
			c.Assert(z, qt.Equals, int32(0))
		})
	}
}

func TestRotation(t *testing.T) {
	for _, tc := range []struct {
		r    GyroRange
		raw  int16
		udps int32
	}{
		{GyroRange2000, 16384, 1000000000},
		{GyroRange2000, -32768, -2000000000},
		{GyroRange1000, 16384, 500000000},
		{GyroRange500, 16384, 250000000},
		{GyroRange250, 16384, 125000000},
		{GyroRange125, 16384, 62500000},
		{GyroRange125, 262, 999450},
	} {
		forEachBus(t, func(c *qt.C, dev *Device, f *fake) {
			c.Assert(dev.Configure(Config{GyroRange: tc.r}), qt.IsNil)
			putInt16(f.regs[RegGyroData:], 0)
			putInt16(f.regs[RegGyroData+2:], tc.raw)
			putInt16(f.regs[RegGyroData+4:], tc.raw)
			x, y, z, err := dev.ReadRotation()
			c.Assert(err, qt.IsNil)
			c.Assert(x, qt.Equals, int32(0))
			c.Assert(y, qt.Equals, tc.udps)
			c.Assert(z, qt.Equals, tc.udps)
		})
	}
}

func TestTemperature(t *testing.T) {
	for _, tc := range []struct {
		raw int16
		mc  int32
	}{
		{0, 23000},
		{512, 24000},
		{-512, 22000},
		{0x7FFF, 86998},
		{-0x7FFF, -40998},
	} {
		forEachBus(t, func(c *qt.C, dev *Device, f *fake) {
			c.Assert(dev.Connected(), qt.IsTrue)
			putInt16(f.regs[RegTemperature:], tc.raw)
			temp, err := dev.ReadTemperature()
			c.Assert(err, qt.IsNil)
			c.Assert(temp, qt.Equals, tc.mc)
		})
	}

	forEachBus(t, func(c *qt.C, dev *Device, f *fake) {
		c.Assert(dev.Connected(), qt.IsTrue)
		_, err := dev.ReadTemperature()
		c.Assert(err, qt.Equals, ErrInvalidTemperature)
	})
}

func TestUpdate(t *testing.T) {
	forEachBus(t, func(c *qt.C, dev *Device, f *fake) {
		c.Assert(dev.Configure(Config{AccelRange: AccelRange2G, GyroRange: GyroRange2000}), qt.IsNil)
		putInt16(f.regs[RegAccelData:], 16384)
		putInt16(f.regs[RegGyroData:], 16384)
		putInt16(f.regs[RegTemperature:], 512)
		c.Assert(dev.Update(drivers.Acceleration|drivers.AngularVelocity|drivers.Temperature), qt.IsNil)
		x, _, _ := dev.Acceleration()
		c.Assert(x, qt.Equals, int32(1000000))
		x, _, _ = dev.AngularVelocity()
		c.Assert(x, qt.Equals, int32(1000000000))
		c.Assert(dev.Temperature(), qt.Equals, int32(24000))

		// Values stay until the next Update.
		putInt16(f.regs[RegAccelData:], 0)
		x, _, _ = dev.Acceleration()
		c.Assert(x, qt.Equals, int32(1000000))

		f.err = errors.New("bus error")
		c.Assert(dev.Update(drivers.Acceleration), qt.Equals, f.err)
		c.Assert(dev.Update(drivers.Temperature), qt.Equals, f.err)
		c.Assert(dev.Update(drivers.Pressure), qt.IsNil)
		_, _, _, err := dev.ReadAcceleration()
		c.Assert(err, qt.Equals, f.err)
		_, _, _, err = dev.ReadRotation()
		c.Assert(err, qt.Equals, f.err)
	})
}

func TestDataReady(t *testing.T) {
	forEachBus(t, func(c *qt.C, dev *Device, f *fake) {
		c.Assert(dev.Connected(), qt.IsTrue)
		f.regs[RegStatus] = statusAccelReady
		accel, gyro, err := dev.DataReady()
		c.Assert(err, qt.IsNil)
		c.Assert(accel, qt.IsTrue)
		c.Assert(gyro, qt.IsFalse)

		f.regs[RegStatus] = statusGyroReady
		accel, gyro, err = dev.DataReady()
		c.Assert(err, qt.IsNil)
		c.Assert(accel, qt.IsFalse)
		c.Assert(gyro, qt.IsTrue)

		f.err = errors.New("bus error")
		_, _, err = dev.DataReady()
		c.Assert(err, qt.Equals, f.err)
	})
}
