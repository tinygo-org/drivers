package ltr390

import (
	"errors"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/tester"
)

func newDevice(c *qt.C) (*Device, *tester.I2CDevice8) {
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice8(c, DefaultAddress)
	fake.Registers[PartID] = 0xB2
	fake.Registers[MainStatus] = statusDataReady
	bus.AddDevice(fake)
	d := New(bus)
	return &d, fake
}

func set20(fake *tester.I2CDevice8, reg uint8, v uint32) {
	fake.Registers[reg] = byte(v)
	fake.Registers[reg+1] = byte(v >> 8)
	fake.Registers[reg+2] = byte(v >> 16)
}

func get20(fake *tester.I2CDevice8, reg uint8) uint32 {
	return uint32(fake.Registers[reg]) | uint32(fake.Registers[reg+1])<<8 | uint32(fake.Registers[reg+2])<<16
}

func TestConnected(t *testing.T) {
	c := qt.New(t)
	d, fake := newDevice(c)
	c.Assert(d.Connected(), qt.IsTrue)

	fake.Registers[PartID] = 0x12
	c.Assert(d.Connected(), qt.IsFalse)

	fake.Registers[PartID] = 0xB2
	fake.Err = errors.New("nack")
	c.Assert(d.Connected(), qt.IsFalse)
}

// Expected register values come from datasheet sections 6 and 9.
func TestConfigure(t *testing.T) {
	cases := []struct {
		name     string
		cfg      Config
		measRate uint8
		gain     uint8
		ctrl     uint8
	}{
		{"defaults", Config{}, 0x22, 0x01, 0x02},
		{"datasheet example", Config{Gain: Gain18, Resolution: Resolution20Bit, Rate: Rate500ms, Mode: ModeUVS}, 0x04, 0x04, 0x0A},
		{"only gain", Config{Gain: Gain18}, 0x22, 0x04, 0x02},
		{"only resolution", Config{Resolution: Resolution13Bit}, 0x52, 0x01, 0x02},
		{"only rate", Config{Rate: Rate2000ms}, 0x26, 0x01, 0x02},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			d, fake := newDevice(c)

			c.Assert(d.Configure(tc.cfg), qt.IsNil)
			c.Assert(fake.Registers[MeasRate], qt.Equals, tc.measRate)
			c.Assert(fake.Registers[GainRange], qt.Equals, tc.gain)
			c.Assert(fake.Registers[MainCtrl], qt.Equals, tc.ctrl)
		})
	}
}

func TestConfigureResetsInterrupt(t *testing.T) {
	c := qt.New(t)
	d, fake := newDevice(c)
	for reg := IntCfg; reg <= ThresLow+2; reg++ {
		fake.Registers[reg] = 0xAA
	}

	c.Assert(d.Configure(Config{}), qt.IsNil)
	c.Assert(fake.Registers[IntCfg], qt.Equals, uint8(0x10))
	c.Assert(fake.Registers[IntPst], qt.Equals, uint8(0x00))
	c.Assert(get20(fake, ThresUp), qt.Equals, uint32(0xFFFFF))
	c.Assert(get20(fake, ThresLow), qt.Equals, uint32(0))
}

func TestSetters(t *testing.T) {
	c := qt.New(t)
	d, fake := newDevice(c)
	c.Assert(d.Configure(Config{}), qt.IsNil)

	c.Assert(d.SetGain(Gain18), qt.IsNil)
	c.Assert(fake.Registers[GainRange], qt.Equals, uint8(0x04))

	c.Assert(d.SetResolution(Resolution20Bit), qt.IsNil)
	c.Assert(fake.Registers[MeasRate], qt.Equals, uint8(0x02))

	c.Assert(d.SetMeasurementRate(Rate500ms), qt.IsNil)
	c.Assert(fake.Registers[MeasRate], qt.Equals, uint8(0x04))
}

// Writing MAIN_CTRL starts a new measurement. Datasheet section 6.1.
func TestSettersRestartMeasurement(t *testing.T) {
	setters := []struct {
		name string
		set  func(d *Device) error
	}{
		{"gain", func(d *Device) error { return d.SetGain(Gain1) }},
		{"resolution", func(d *Device) error { return d.SetResolution(Resolution16Bit) }},
		{"rate", func(d *Device) error { return d.SetMeasurementRate(Rate1000ms) }},
	}
	for _, s := range setters {
		t.Run(s.name, func(t *testing.T) {
			c := qt.New(t)
			d, fake := newDevice(c)
			c.Assert(d.Configure(Config{Mode: ModeUVS}), qt.IsNil)

			fake.Registers[MainCtrl] = 0
			c.Assert(s.set(d), qt.IsNil)
			c.Assert(fake.Registers[MainCtrl], qt.Equals, uint8(ctrlEnable|ctrlUVSMode))

			c.Assert(d.Disable(), qt.IsNil)
			c.Assert(s.set(d), qt.IsNil)
			c.Assert(fake.Registers[MainCtrl], qt.Equals, uint8(ctrlUVSMode))
		})
	}
}

func TestEnableDisable(t *testing.T) {
	c := qt.New(t)
	d, fake := newDevice(c)
	c.Assert(d.Configure(Config{Mode: ModeUVS}), qt.IsNil)

	c.Assert(d.Disable(), qt.IsNil)
	c.Assert(fake.Registers[MainCtrl], qt.Equals, uint8(ctrlUVSMode))

	c.Assert(d.Enable(), qt.IsNil)
	c.Assert(fake.Registers[MainCtrl], qt.Equals, uint8(ctrlEnable|ctrlUVSMode))

	c.Assert(d.Disable(), qt.IsNil)
	_, err := d.ReadRawUVS()
	c.Assert(err, qt.IsNil)
	c.Assert(fake.Registers[MainCtrl], qt.Equals, uint8(ctrlEnable|ctrlUVSMode))
}

func TestSetThresholds(t *testing.T) {
	c := qt.New(t)
	d, fake := newDevice(c)

	c.Assert(d.SetThresholds(0x12345, 0xABCDE), qt.IsNil)
	c.Assert(fake.Registers[ThresLow:ThresLow+3], qt.DeepEquals, []uint8{0x45, 0x23, 0x01})
	c.Assert(fake.Registers[ThresUp:ThresUp+3], qt.DeepEquals, []uint8{0xDE, 0xBC, 0x0A})

	c.Assert(d.SetThresholds(0, 0xFFFFFFFF), qt.IsNil)
	c.Assert(get20(fake, ThresUp), qt.Equals, uint32(0xFFFFF))
}

// Expected register values come from datasheet sections 6.8, 6.9 and 9.
func TestConfigureInterrupt(t *testing.T) {
	cases := []struct {
		name        string
		enable      bool
		source      Mode
		persistence uint8
		intCfg      uint8
		intPst      uint8
	}{
		{"ALS", true, ModeALS, 0, 0x14, 0x00},
		{"UVS", true, ModeUVS, 1, 0x34, 0x10},
		{"disabled", false, ModeUVS, 15, 0x30, 0xF0},
		{"persistence limit", true, ModeALS, 200, 0x14, 0xF0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			d, fake := newDevice(c)

			c.Assert(d.ConfigureInterrupt(tc.enable, tc.source, tc.persistence), qt.IsNil)
			c.Assert(fake.Registers[IntCfg], qt.Equals, tc.intCfg)
			c.Assert(fake.Registers[IntPst], qt.Equals, tc.intPst)
		})
	}
}

func TestDataReady(t *testing.T) {
	cases := []struct {
		status uint8
		want   bool
	}{
		{0x08, true},
		{0x28, true},
		{0x20, false},
		{0x10, false},
		{0x00, false},
	}
	for _, tc := range cases {
		c := qt.New(t)
		d, fake := newDevice(c)
		fake.Registers[MainStatus] = tc.status

		ready, err := d.DataReady()
		c.Assert(err, qt.IsNil)
		c.Assert(ready, qt.Equals, tc.want, qt.Commentf("status %#x", tc.status))
	}
}

func TestReadRaw(t *testing.T) {
	c := qt.New(t)
	d, fake := newDevice(c)
	c.Assert(d.Configure(Config{}), qt.IsNil)

	set20(fake, ALSData, 0x12345)
	set20(fake, UVSData, 0x6789A)
	// Bits 7:4 of the high byte are reserved. Datasheet section 6.6.
	fake.Registers[ALSData+2] |= 0xF0

	als, err := d.ReadRawALS()
	c.Assert(err, qt.IsNil)
	c.Assert(als, qt.Equals, uint32(0x12345))
	c.Assert(fake.Registers[MainCtrl], qt.Equals, uint8(ctrlEnable))

	uvs, err := d.ReadRawUVS()
	c.Assert(err, qt.IsNil)
	c.Assert(uvs, qt.Equals, uint32(0x6789A))
	c.Assert(fake.Registers[MainCtrl], qt.Equals, uint8(ctrlEnable|ctrlUVSMode))
}

// Expected values follow datasheet section 7.1 with WFAC = 1.
// 13 bit is not in the datasheet table and uses INT = 2^(13-18).
func TestReadIlluminance(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		als  uint32
		want uint32
	}{
		{"gain 3 18 bit", Config{}, 5000, 1000000},
		{"gain 1 16 bit", Config{Gain: Gain1, Resolution: Resolution16Bit}, 1000, 2400000},
		{"gain 6 17 bit", Config{Gain: Gain6, Resolution: Resolution17Bit}, 5000, 1000000},
		{"gain 9 19 bit", Config{Gain: Gain9, Resolution: Resolution19Bit}, 30000, 1000000},
		{"gain 18 20 bit", Config{Gain: Gain18, Resolution: Resolution20Bit}, 120000, 1000000},
		{"gain 3 13 bit", Config{Resolution: Resolution13Bit}, 100, 640000},
		{"20 bit full scale", Config{Gain: Gain1, Resolution: Resolution20Bit}, 0xFFFFF, 157286250},
		{"13 bit full scale", Config{Gain: Gain1, Resolution: Resolution13Bit}, 0x1FFF, 157267200},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			d, fake := newDevice(c)
			c.Assert(d.Configure(tc.cfg), qt.IsNil)
			set20(fake, ALSData, tc.als)

			mlux, err := d.ReadIlluminance()
			c.Assert(err, qt.IsNil)
			c.Assert(mlux, qt.Equals, tc.want)
		})
	}
}

// Expected values follow datasheet sections 4.5 and 7.2 with WFAC = 1.
func TestReadUVIndex(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		uvs  uint32
		want uint32
	}{
		{"gain 18 20 bit", Config{Gain: Gain18, Resolution: Resolution20Bit}, 2300, 1000},
		{"gain 18 20 bit UVI 10", Config{Gain: Gain18, Resolution: Resolution20Bit}, 23000, 10000},
		{"gain 9 19 bit", Config{Gain: Gain9, Resolution: Resolution19Bit}, 575, 1000},
		{"gain 6 17 bit", Config{Gain: Gain6, Resolution: Resolution17Bit}, 575, 6000},
		{"gain 3 18 bit", Config{}, 575, 6000},
		{"gain 1 16 bit", Config{Gain: Gain1, Resolution: Resolution16Bit}, 23, 2880},
		{"gain 18 13 bit", Config{Gain: Gain18, Resolution: Resolution13Bit}, 128, 7123},
		{"20 bit full scale", Config{Gain: Gain1, Resolution: Resolution20Bit}, 0xFFFFF, 8206239},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			d, fake := newDevice(c)
			c.Assert(d.Configure(tc.cfg), qt.IsNil)
			set20(fake, UVSData, tc.uvs)

			uvi, err := d.ReadUVIndex()
			c.Assert(err, qt.IsNil)
			c.Assert(uvi, qt.Equals, tc.want)
		})
	}
}

// New uses the power-on defaults, so reads work without Configure.
func TestReadWithoutConfigure(t *testing.T) {
	c := qt.New(t)
	d, fake := newDevice(c)
	c.Assert(d.Address, qt.Equals, uint16(DefaultAddress))
	set20(fake, ALSData, 5000)

	mlux, err := d.ReadIlluminance()
	c.Assert(err, qt.IsNil)
	c.Assert(mlux, qt.Equals, uint32(1000000))
	c.Assert(fake.Registers[MainCtrl], qt.Equals, uint8(ctrlEnable))
}

func TestReadTimeout(t *testing.T) {
	c := qt.New(t)
	d, fake := newDevice(c)
	c.Assert(d.Configure(Config{Resolution: Resolution13Bit, Rate: Rate25ms}), qt.IsNil)
	fake.Registers[MainStatus] = 0

	_, err := d.ReadRawALS()
	c.Assert(err, qt.Equals, ErrTimeout)
}

func TestBusError(t *testing.T) {
	c := qt.New(t)
	d, fake := newDevice(c)
	errNack := errors.New("nack")
	fake.Err = errNack

	c.Assert(d.Configure(Config{}), qt.Equals, errNack)
	_, err := d.ReadIlluminance()
	c.Assert(err, qt.Equals, errNack)
	_, err = d.ReadUVIndex()
	c.Assert(err, qt.Equals, errNack)
}

// The period is the longer of conversion time and measurement rate. Datasheet section 6.2.
func TestPeriod(t *testing.T) {
	cases := []struct {
		res  Resolution
		rate MeasurementRate
		want time.Duration
	}{
		{Resolution18Bit, Rate100ms, 100 * time.Millisecond},
		{Resolution20Bit, Rate25ms, 400 * time.Millisecond},
		{Resolution19Bit, Rate100ms, 200 * time.Millisecond},
		{Resolution20Bit, Rate2000ms, 2000 * time.Millisecond},
		{Resolution16Bit, Rate25ms, 25 * time.Millisecond},
		{Resolution13Bit, Rate25ms, 25 * time.Millisecond},
	}
	for _, tc := range cases {
		c := qt.New(t)
		d := Device{resolution: tc.res, rate: tc.rate}
		c.Assert(d.period(), qt.Equals, tc.want)
	}
}
