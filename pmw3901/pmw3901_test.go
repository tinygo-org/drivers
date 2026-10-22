package pmw3901

import (
	"errors"
	"reflect"
	"testing"

	qt "github.com/frankban/quicktest"
)

var errBus = errors.New("bus error")

type regValue struct {
	reg, val uint8
}

func pairs(regs string) []regValue {
	var r []regValue
	for i := 0; i < len(regs); i += 2 {
		r = append(r, regValue{regs[i], regs[i+1]})
	}
	return r
}

// fake emulates the PMW3901 SPI protocol. A transaction is the bytes sent
// while cs is low.
type fake struct {
	c      *qt.C
	regs   [128]byte
	burst  [12]byte
	csLow  bool
	tx     []byte // bytes of the current transaction
	writes []regValue
	resets int
	ops    int
	failAt int // fail the Nth Tx when non-zero
	err    error
	id     uint8

	statusPolls int    // RawData_Grab_Status reads before ready, -1 for never
	frame       []byte // pixels to stream from RawData_Grab
	stream      []byte // pending RawData_Grab values
}

func newFake(c *qt.C) *fake {
	f := &fake{c: c, id: ProductID}
	f.reset()
	return f
}

func (f *fake) reset() {
	f.regs = [128]byte{}
	f.regs[RegProductID] = f.id
	f.regs[RegInverseProductID] = InverseProductID
}

func (f *fake) Set(level bool) {
	if !level {
		f.c.Assert(f.csLow, qt.IsFalse, qt.Commentf("cs already low"))
		f.csLow = true
		f.tx = f.tx[:0]
		return
	}
	if f.csLow && len(f.tx) > 0 && f.tx[0]&spiWrite != 0 {
		f.c.Assert(f.tx, qt.HasLen, 2)
		f.write(f.tx[0]&^spiWrite, f.tx[1])
	}
	f.csLow = false
}

func (f *fake) write(reg, val uint8) {
	f.writes = append(f.writes, regValue{reg, val})
	if reg == RegPowerUpReset && val == powerUpResetValue {
		f.resets++
		f.reset()
		return
	}
	f.regs[reg] = val
	if reg == RegRawDataGrab && val == 0x00 {
		f.c.Assert(f.regs[RegRawDataGrabStat], qt.Not(qt.Equals), uint8(0))
		// Invalid values and unused bits must be ignored.
		f.stream = []byte{0x00, 0xC0}
		for _, p := range f.frame {
			f.stream = append(f.stream, rawDataUpper|p>>2, 0x00, rawDataLower|(p&0x03)<<2|0x33)
		}
	}
}

func (f *fake) readReg(reg uint8) uint8 {
	switch reg {
	case RegRawDataGrabStat:
		if f.regs[RegRawDataGrab] == 0xFF && f.statusPolls >= 0 {
			if f.statusPolls == 0 {
				f.regs[RegRawDataGrabStat] = 0xC0
			}
			f.statusPolls--
		}
	case RegRawDataGrab:
		if len(f.stream) == 0 {
			return 0x00
		}
		v := f.stream[0]
		f.stream = f.stream[1:]
		return v
	}
	return f.regs[reg]
}

func (f *fake) Tx(w, r []byte) error {
	f.c.Assert(f.csLow, qt.IsTrue, qt.Commentf("cs not low"))
	f.ops++
	if f.failAt != 0 && f.ops == f.failAt {
		return errBus
	}
	if f.err != nil {
		return f.err
	}
	if r != nil {
		f.c.Assert(len(r), qt.Equals, len(w))
	}
	for i, b := range w {
		pos := len(f.tx)
		f.tx = append(f.tx, b)
		if pos == 0 || f.tx[0]&spiWrite != 0 {
			continue
		}
		f.c.Assert(r, qt.Not(qt.IsNil))
		reg := f.tx[0]
		if reg == RegMotionBurst {
			f.c.Assert(pos <= len(f.burst), qt.IsTrue)
			r[i] = f.burst[pos-1]
		} else {
			f.c.Assert(pos, qt.Equals, 1)
			r[i] = f.readReg(reg)
		}
	}
	return nil
}

func (f *fake) Transfer(b byte) (byte, error) {
	f.c.Fatalf("unexpected Transfer")
	return 0, nil
}

func assertWrites(c *qt.C, got, want []regValue) {
	c.Helper()
	if !reflect.DeepEqual(got, want) {
		c.Fatalf("writes\ngot  %x\nwant %x", got, want)
	}
}

func newDevice(c *qt.C) (*Device, *fake) {
	f := newFake(c)
	dev := New(f, f)
	return &dev, f
}

func TestInitTables(t *testing.T) {
	c := qt.New(t)
	c.Assert(pairs(initRegisters1), qt.HasLen, 59)
	c.Assert(pairs(initRegisters2), qt.HasLen, 14)
	c.Assert(initRegisters2[len(initRegisters2)-2:], qt.Equals, "\x40\x80")
}

func TestConnected(t *testing.T) {
	c := qt.New(t)
	dev, f := newDevice(c)
	c.Assert(dev.Connected(), qt.IsTrue)

	f.regs[RegInverseProductID] = 0
	c.Assert(dev.Connected(), qt.IsFalse)

	f.regs[RegInverseProductID] = InverseProductID
	f.regs[RegProductID] = 0
	c.Assert(dev.Connected(), qt.IsFalse)

	f.regs[RegProductID] = ProductID
	f.err = errBus
	c.Assert(dev.Connected(), qt.IsFalse)
	c.Assert(f.csLow, qt.IsFalse)
}

func TestConfigure(t *testing.T) {
	c := qt.New(t)
	dev, f := newDevice(c)
	c.Assert(dev.Configure(), qt.IsNil)
	c.Assert(f.resets, qt.Equals, 1)
	c.Assert(f.csLow, qt.IsFalse)

	want := []regValue{{RegPowerUpReset, powerUpResetValue}}
	want = append(want, pairs(initRegisters1)...)
	want = append(want, pairs(initRegisters2)...)
	assertWrites(c, f.writes, want)
}

func TestConfigureNotConnected(t *testing.T) {
	c := qt.New(t)
	dev, f := newDevice(c)
	f.id = 0x48
	c.Assert(dev.Configure(), qt.Equals, ErrDeviceNotConnected)
	assertWrites(c, f.writes, []regValue{{RegPowerUpReset, powerUpResetValue}})
}

func TestConfigureBusErrorAtEachStep(t *testing.T) {
	c := qt.New(t)
	dev, f := newDevice(c)
	for n := 1; ; n++ {
		f.ops = 0
		f.failAt = n
		err := dev.Configure()
		c.Assert(f.csLow, qt.IsFalse)
		if f.ops < n {
			c.Assert(err, qt.IsNil)
			break
		}
		if err != errBus && err != ErrDeviceNotConnected {
			c.Fatalf("step %d: unexpected error %v", n, err)
		}
	}
}

func TestReadMotion(t *testing.T) {
	c := qt.New(t)
	dev, f := newDevice(c)
	f.regs[RegMotion] = motionMOT
	f.regs[RegDeltaXL] = 0x34
	f.regs[RegDeltaXH] = 0x12
	f.regs[RegDeltaYL] = 0xFE
	f.regs[RegDeltaYH] = 0xFF
	dx, dy, err := dev.ReadMotion()
	c.Assert(err, qt.IsNil)
	c.Assert(dx, qt.Equals, int16(0x1234))
	c.Assert(dy, qt.Equals, int16(-2))

	for n := 1; n <= 10; n++ {
		f.ops = 0
		f.failAt = n
		_, _, err = dev.ReadMotion()
		c.Assert(err, qt.Equals, errBus)
		c.Assert(f.csLow, qt.IsFalse)
	}
}

func TestReadMotionBurst(t *testing.T) {
	c := qt.New(t)
	dev, f := newDevice(c)
	f.burst = [12]byte{0xB0, 0x3F, 0x05, 0x00, 0xFB, 0xFF, 0x60, 0x20, 0x80, 0x10, 0x01, 0x02}
	m, err := dev.ReadMotionBurst()
	c.Assert(err, qt.IsNil)
	c.Assert(m, qt.Equals, Motion{
		Motion:      0xB0,
		Observation: 0x3F,
		DeltaX:      5,
		DeltaY:      -5,
		Squal:       0x60,
		RawDataSum:  0x20,
		MaxRawData:  0x80,
		MinRawData:  0x10,
		Shutter:     0x0102,
	})
	c.Assert(m.Valid(), qt.IsTrue)

	for n := 1; n <= 2; n++ {
		f.ops = 0
		f.failAt = n
		m, err = dev.ReadMotionBurst()
		c.Assert(err, qt.Equals, errBus)
		c.Assert(m, qt.Equals, Motion{})
		c.Assert(f.csLow, qt.IsFalse)
	}
}

func TestMotionValid(t *testing.T) {
	c := qt.New(t)
	c.Assert(Motion{Motion: motionMOT, Squal: 0x19, Shutter: 0x1F00}.Valid(), qt.IsTrue)
	c.Assert(Motion{Motion: motionMOT, Squal: 0x18, Shutter: 0x1E00}.Valid(), qt.IsTrue)
	c.Assert(Motion{Motion: motionMOT, Squal: 0x18, Shutter: 0x1F00}.Valid(), qt.IsFalse)
	c.Assert(Motion{Motion: 0, Squal: 0x60, Shutter: 0x0100}.Valid(), qt.IsFalse)
}

func TestSetLED(t *testing.T) {
	c := qt.New(t)
	dev, f := newDevice(c)
	c.Assert(dev.SetLED(true), qt.IsNil)
	c.Assert(dev.SetLED(false), qt.IsNil)
	assertWrites(c, f.writes, []regValue{
		{regBank, 0x14}, {0x6F, 0x1C}, {regBank, 0x00},
		{regBank, 0x14}, {0x6F, 0x00}, {regBank, 0x00},
	})

	for n := 1; n <= 3; n++ {
		f.ops = 0
		f.failAt = n
		c.Assert(dev.SetLED(true), qt.Equals, errBus)
		c.Assert(f.csLow, qt.IsFalse)
	}
}

func TestRegisterAccess(t *testing.T) {
	c := qt.New(t)
	dev, f := newDevice(c)
	c.Assert(dev.WriteRegister(0x55, 0xAA), qt.IsNil)
	c.Assert(f.regs[0x55], qt.Equals, uint8(0xAA))
	val, err := dev.ReadRegister(0x55)
	c.Assert(err, qt.IsNil)
	c.Assert(val, qt.Equals, uint8(0xAA))

	f.err = errBus
	c.Assert(dev.WriteRegister(0x55, 0), qt.Equals, errBus)
	_, err = dev.ReadRegister(0x55)
	c.Assert(err, qt.Equals, errBus)
	c.Assert(f.csLow, qt.IsFalse)
}

func TestReadFrame(t *testing.T) {
	c := qt.New(t)
	dev, f := newDevice(c)
	f.statusPolls = 3
	f.frame = make([]byte, FrameSize)
	for i := range f.frame {
		f.frame[i] = byte(i * 7)
	}
	var frame [FrameSize]byte
	for i := range frame {
		frame[i] = 0xFF
	}
	c.Assert(dev.ReadFrame(&frame), qt.IsNil)
	c.Assert(frame[:], qt.DeepEquals, f.frame)
	c.Assert(f.csLow, qt.IsFalse)

	want := pairs(frameRegisters1)
	want = append(want, pairs(frameRegisters2)...)
	want = append(want, regValue{RegRawDataGrab, 0x00})
	assertWrites(c, f.writes, want)
}

func TestReadFrameErrors(t *testing.T) {
	t.Run("status timeout", func(t *testing.T) {
		c := qt.New(t)
		dev, f := newDevice(c)
		f.statusPolls = -1
		var frame [FrameSize]byte
		c.Assert(dev.ReadFrame(&frame), qt.Equals, ErrTimeout)
	})

	t.Run("pixel timeout", func(t *testing.T) {
		c := qt.New(t)
		dev, f := newDevice(c)
		f.frame = make([]byte, FrameSize/2)
		var frame [FrameSize]byte
		c.Assert(dev.ReadFrame(&frame), qt.Equals, ErrTimeout)
	})

	t.Run("bus error", func(t *testing.T) {
		c := qt.New(t)
		dev, f := newDevice(c)
		f.frame = make([]byte, FrameSize)
		var frame [FrameSize]byte
		// Steps 1 to 15 are the setup writes, the status read, the grab write and the first pixel read.
		for n := 1; n <= 15; n++ {
			f.reset()
			f.statusPolls = 0
			f.writes = nil
			f.ops = 0
			f.failAt = n
			c.Assert(dev.ReadFrame(&frame), qt.Equals, errBus, qt.Commentf("step %d", n))
			c.Assert(f.csLow, qt.IsFalse)
		}
	})
}
