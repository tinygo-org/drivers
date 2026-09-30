package ads1015

import (
	"errors"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/tester"
)

// newFakeDevice wires a Device to an in-memory I2C bus with one fake
// ADS1015 at the default address. The register map starts in the
// power-on-like state: idle (OS/ready bit set) and zeroed thresholds.
func newFakeDevice(c *qt.C) (*Device, *tester.I2CDevice16) {
	bus := tester.NewI2CBus(c)
	fake := tester.NewI2CDevice16(c, uint8(Address))
	fake.Registers = map[uint8]uint16{
		regConversion:    0,
		regConfig:        configOSReady,
		regLowThreshold:  0,
		regHighThreshold: 0,
	}
	bus.AddDevice(fake)

	return New(bus), fake
}

// TestNew checks that New() stores the given bus/address and precomputes
// conversionPollInterval from DefaultConfig's DataRate, since Read() relies
// on that field rather than recomputing it on every call.
func TestNew(t *testing.T) {
	c := qt.New(t)
	dev, _ := newFakeDevice(c)

	c.Assert(dev.Address, qt.Equals, uint16(Address))

	got := dev.Config()
	c.Assert(got.Gain, qt.Equals, DefaultConfig.Gain)
	c.Assert(got.Mode, qt.Equals, DefaultConfig.Mode)
	c.Assert(got.DataRate, qt.Equals, DefaultConfig.DataRate)
	c.Assert(got.conversionPollInterval, qt.Equals, ConversionDuration(DefaultConfig.DataRate))
}

// TestConfigure checks that Configure() stores the given config, derives
// conversionPollInterval from its DataRate (issue 1), and starts a
// conversion on channel 0 by writing the config register accordingly.
func TestConfigure(t *testing.T) {
	c := qt.New(t)
	dev, fake := newFakeDevice(c)

	cfg := Config{
		Gain:               Gain4096mV,
		Mode:               ModeContinuous,
		DataRate:           DataRate250SPS,
		ComparatorMode:     ComparatorModeWindow,
		ComparatorPolarity: ComparatorPolarityActiveHigh,
		ComparatorLatch:    ComparatorLatching,
		ComparatorQueue:    ComparatorQueueAfter2Conv,
	}
	err := dev.Configure(cfg)
	c.Assert(err, qt.IsNil)

	got := dev.Config()
	c.Assert(got.Gain, qt.Equals, cfg.Gain)
	c.Assert(got.DataRate, qt.Equals, cfg.DataRate)
	c.Assert(got.conversionPollInterval, qt.Equals, ConversionDuration(DataRate250SPS))

	want := configOSStart | uint16(MuxSingle0) | uint16(cfg.Gain) | uint16(cfg.Mode) |
		uint16(cfg.DataRate) | uint16(cfg.ComparatorMode) | uint16(cfg.ComparatorPolarity) |
		uint16(cfg.ComparatorLatch) | uint16(cfg.ComparatorQueue)
	c.Assert(fake.Registers[regConfig], qt.Equals, want)
}

// TestRequest checks that Request() (and thus Read(), which shares the same
// startConversion() call) assembles the config register from every Config
// field plus the mux argument, with the OS/start bit set to trigger a new
// conversion. This is the "config word that startConversion() writes"
// coverage called out in issue 8.
func TestRequest(t *testing.T) {
	c := qt.New(t)
	dev, fake := newFakeDevice(c)

	cfg := Config{
		Gain:               Gain0512mV,
		Mode:               ModeSingle,
		DataRate:           DataRate3300SPS,
		ComparatorMode:     ComparatorModeTraditional,
		ComparatorPolarity: ComparatorPolarityActiveLow,
		ComparatorLatch:    ComparatorNonLatching,
		ComparatorQueue:    ComparatorQueueAfter4Conv,
	}
	c.Assert(dev.Configure(cfg), qt.IsNil)

	err := dev.Request(MuxDiff13)
	c.Assert(err, qt.IsNil)

	want := configOSStart | uint16(MuxDiff13) | uint16(cfg.Gain) | uint16(cfg.Mode) |
		uint16(cfg.DataRate) | uint16(cfg.ComparatorMode) | uint16(cfg.ComparatorPolarity) |
		uint16(cfg.ComparatorLatch) | uint16(cfg.ComparatorQueue)
	c.Assert(fake.Registers[regConfig], qt.Equals, want)
}

// TestValue checks the sign extension in Value(): the ADS1015 left-justifies
// its 12-bit result in the 16-bit conversion register, so Value() must
// arithmetic-shift right by 4 to recover a signed 12-bit code, not just mask
// off the low bits.
func TestValue(t *testing.T) {
	cases := []struct {
		name string
		raw  uint16
		want int16
	}{
		{"positive full scale", 0x7FF0, 2047},
		{"negative full scale", 0x8000, -2048},
		{"minus one", 0xFFF0, -1},
		{"small positive", 0x0010, 1},
		{"zero", 0x0000, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			dev, fake := newFakeDevice(c)
			fake.Registers[regConversion] = tc.raw

			got, err := dev.Value()
			c.Assert(err, qt.IsNil)
			c.Assert(got, qt.Equals, tc.want)
		})
	}
}

// TestToVoltage checks the raw-code-to-millivolt conversion for each gain
// setting. It uses codes at +/-maxCode so the expected result is an exact,
// hand-computable value, which also documents issue 3: a raw reading of
// +2047 (not +2048) maps to the full-scale voltage, because ToVoltage
// divides by maxCode (2047) rather than the datasheet's 2048.
func TestToVoltage(t *testing.T) {
	cases := []struct {
		gain Gain
		raw  int16
		want int32
	}{
		{Gain6144mV, 2047, 6144},
		{Gain6144mV, -2047, -6144},
		{Gain4096mV, 2047, 4096},
		{Gain2048mV, 2047, 2048},
		{Gain1024mV, 2047, 1024},
		{Gain0512mV, 2047, 512},
		{Gain0256mV, 2047, 256},
		{Gain6144mV, 0, 0},
	}
	for _, tc := range cases {
		c := qt.New(t)
		dev, _ := newFakeDevice(c)
		cfg := dev.Config()
		cfg.Gain = tc.gain
		c.Assert(dev.Configure(cfg), qt.IsNil)

		c.Assert(dev.ToVoltage(tc.raw), qt.Equals, tc.want)
	}
}

// TestReady checks that Ready() reflects the OS bit of the config register:
// clear while a conversion is in progress, set once it completes.
func TestReady(t *testing.T) {
	c := qt.New(t)
	dev, fake := newFakeDevice(c)

	fake.Registers[regConfig] = 0x0000
	ready, err := dev.Ready()
	c.Assert(err, qt.IsNil)
	c.Assert(ready, qt.IsFalse)

	fake.Registers[regConfig] = configOSReady
	ready, err = dev.Ready()
	c.Assert(err, qt.IsNil)
	c.Assert(ready, qt.IsTrue)
}

// TestConnected checks that Connected() turns an I2C error from reading the
// config register into false, and a successful read into true.
func TestConnected(t *testing.T) {
	c := qt.New(t)
	dev, fake := newFakeDevice(c)
	c.Assert(dev.Connected(), qt.IsTrue)

	fake.Err = errors.New("nack")
	c.Assert(dev.Connected(), qt.IsFalse)
}

// TestThresholds checks that SetThresholdLow/SetThresholdHigh write the raw
// 16-bit value given, and ThresholdLow/ThresholdHigh read it back
// reinterpreted as signed (needed since the comparator thresholds are
// signed 12-bit codes left-justified the same way as the conversion
// register).
func TestThresholds(t *testing.T) {
	c := qt.New(t)
	dev, fake := newFakeDevice(c)

	c.Assert(dev.SetThresholdLow(0xFF00), qt.IsNil)
	low, err := dev.ThresholdLow()
	c.Assert(err, qt.IsNil)
	c.Assert(low, qt.Equals, int16(-256))
	c.Assert(fake.Registers[regLowThreshold], qt.Equals, uint16(0xFF00))

	c.Assert(dev.SetThresholdHigh(0x7F00), qt.IsNil)
	high, err := dev.ThresholdHigh()
	c.Assert(err, qt.IsNil)
	c.Assert(high, qt.Equals, int16(0x7F00))
	c.Assert(fake.Registers[regHighThreshold], qt.Equals, uint16(0x7F00))
}

// TestEnableConversionReadyPin checks that it writes the documented
// "conversion ready" sentinel values (high threshold MSB set, low threshold
// MSB clear, issue 7's magic-value constants) to the threshold registers,
// and that it only promotes ComparatorQueue off ComparatorQueueDisable, so
// it never clobbers a queue depth the caller already chose.
func TestEnableConversionReadyPin(t *testing.T) {
	t.Run("disabled queue is promoted", func(t *testing.T) {
		c := qt.New(t)
		dev, fake := newFakeDevice(c)

		err := dev.EnableConversionReadyPin()
		c.Assert(err, qt.IsNil)
		c.Assert(fake.Registers[regHighThreshold], qt.Equals, conversionReadyHiThresh)
		c.Assert(fake.Registers[regLowThreshold], qt.Equals, conversionReadyLoThresh)
		c.Assert(dev.Config().ComparatorQueue, qt.Equals, ComparatorQueueAfter1Conv)
	})

	t.Run("existing queue depth is preserved", func(t *testing.T) {
		c := qt.New(t)
		dev, fake := newFakeDevice(c)
		cfg := dev.Config()
		cfg.ComparatorQueue = ComparatorQueueAfter4Conv
		c.Assert(dev.Configure(cfg), qt.IsNil)

		err := dev.EnableConversionReadyPin()
		c.Assert(err, qt.IsNil)
		c.Assert(fake.Registers[regHighThreshold], qt.Equals, conversionReadyHiThresh)
		c.Assert(fake.Registers[regLowThreshold], qt.Equals, conversionReadyLoThresh)
		c.Assert(dev.Config().ComparatorQueue, qt.Equals, ComparatorQueueAfter4Conv)
	})
}

// TestReadADCInvalidChannel checks that channel numbers outside 0-3 are
// rejected before touching the bus, for both the blocking and
// request/response APIs.
func TestReadADCInvalidChannel(t *testing.T) {
	c := qt.New(t)
	dev, _ := newFakeDevice(c)

	_, err := dev.ReadADC(4)
	c.Assert(err, qt.Equals, ErrInvalidChannel)

	err = dev.RequestADC(4)
	c.Assert(err, qt.Equals, ErrInvalidChannel)
}

// TestReadVoltage checks that ReadVoltage combines a conversion on the
// requested channel with ToVoltage's gain-scaled conversion.
func TestReadVoltage(t *testing.T) {
	c := qt.New(t)
	dev, fake := newFakeDevice(c)
	cfg := dev.Config()
	cfg.Gain = Gain2048mV
	cfg.DataRate = DataRate3300SPS
	c.Assert(dev.Configure(cfg), qt.IsNil)

	// Full-scale positive code, left-justified.
	fake.Registers[regConversion] = 0x7FF0

	mv, err := dev.ReadVoltage(0)
	c.Assert(err, qt.IsNil)
	c.Assert(mv, qt.Equals, int32(2048))
	c.Assert(fake.Registers[regConfig]&uint16(MuxSingle0), qt.Equals, uint16(MuxSingle0))
}

// TestReadContinuousModeWaitsForDataRate is a regression test for issue 1:
// in continuous mode, Read() must wait at least one full conversion period
// at the configured DataRate before reading Value(), or it can return the
// previous mux setting's stale result. Before the fix this only slept a
// fixed 1ms, which is far shorter than the ~7.8ms a 128 SPS conversion
// takes.
func TestReadContinuousModeWaitsForDataRate(t *testing.T) {
	c := qt.New(t)
	dev, fake := newFakeDevice(c)

	cfg := dev.Config()
	cfg.Mode = ModeContinuous
	cfg.DataRate = DataRate128SPS
	c.Assert(dev.Configure(cfg), qt.IsNil)

	fake.Registers[regConversion] = 0x0100

	start := time.Now()
	val, err := dev.Read(MuxSingle1)
	elapsed := time.Since(start)

	c.Assert(err, qt.IsNil)
	c.Assert(val, qt.Equals, int16(0x0100)>>4)
	c.Assert(elapsed >= ConversionDuration(DataRate128SPS), qt.IsTrue,
		qt.Commentf("Read() only waited %s, want at least %s", elapsed, ConversionDuration(DataRate128SPS)))
}

// TestReadSingleShotWaitsBeforeFirstPoll is a regression test for issue 4:
// Read() must sleep before its first Ready() poll in single-shot mode. The
// fake device reports "ready" immediately (its OS bit is never cleared by a
// real conversion), which reproduces the worst case where the very first
// poll would misreport completion; the elapsed time must still cover one
// full conversion period, proving the pre-poll sleep executed.
func TestReadSingleShotWaitsBeforeFirstPoll(t *testing.T) {
	c := qt.New(t)
	dev, fake := newFakeDevice(c)

	cfg := dev.Config()
	cfg.Mode = ModeSingle
	cfg.DataRate = DataRate128SPS
	c.Assert(dev.Configure(cfg), qt.IsNil)

	// The fake never clears the OS bit, so without a pre-poll sleep,
	// Read() would return almost instantly.
	fake.Registers[regConfig] = configOSReady
	fake.Registers[regConversion] = 0x0200

	start := time.Now()
	val, err := dev.Read(MuxSingle2)
	elapsed := time.Since(start)

	c.Assert(err, qt.IsNil)
	c.Assert(val, qt.Equals, int16(0x0200)>>4)
	c.Assert(elapsed >= ConversionDuration(DataRate128SPS), qt.IsTrue,
		qt.Commentf("Read() returned after only %s, want at least %s", elapsed, ConversionDuration(DataRate128SPS)))
}

// stuckBus is a drivers.I2C fake whose ADS1015 emulation clears the OS/ready
// bit whenever the config register is written with the start bit set, and
// never sets it again, as if a conversion had stalled. tester.I2CDevice16's
// plain register map can't express this: startConversion's own write always
// includes the OS/start bit (0x8000), and since that bit shares its
// position with the OS/ready bit read back, a static map immediately reads
// back as "ready".
type stuckBus struct {
	config uint16
}

func (b *stuckBus) Tx(addr uint16, w, r []byte) error {
	switch {
	case len(w) == 1 && r != nil:
		var val uint16
		if w[0] == regConfig {
			val = b.config
		}
		r[0], r[1] = byte(val>>8), byte(val)
	case len(w) == 3:
		val := uint16(w[1])<<8 | uint16(w[2])
		if w[0] == regConfig {
			b.config = val &^ configOSReady
		}
	}
	return nil
}

// TestReadTimeout checks that Read() gives up and returns ErrTimeout in
// single-shot mode if the OS bit never sets, instead of polling forever.
func TestReadTimeout(t *testing.T) {
	c := qt.New(t)
	dev := New(&stuckBus{})

	cfg := dev.Config()
	cfg.Mode = ModeSingle
	cfg.DataRate = DataRate3300SPS
	c.Assert(dev.Configure(cfg), qt.IsNil)

	_, err := dev.Read(MuxSingle0)
	c.Assert(err, qt.Equals, ErrTimeout)
}

// TestConversionDuration checks the datasheet-derived period for every
// DataRate value, plus the fallback for an unrecognized value.
func TestConversionDuration(t *testing.T) {
	cases := []struct {
		rate DataRate
		want time.Duration
	}{
		{DataRate128SPS, 7813 * time.Microsecond},
		{DataRate250SPS, 4000 * time.Microsecond},
		{DataRate490SPS, 2041 * time.Microsecond},
		{DataRate920SPS, 1087 * time.Microsecond},
		{DataRate1600SPS, 625 * time.Microsecond},
		{DataRate2400SPS, 417 * time.Microsecond},
		{DataRate3300SPS, 303 * time.Microsecond},
		{DataRate(0xFFFF), 8 * time.Millisecond},
	}
	for _, tc := range cases {
		c := qt.New(t)
		c.Assert(ConversionDuration(tc.rate), qt.Equals, tc.want)
	}
}
