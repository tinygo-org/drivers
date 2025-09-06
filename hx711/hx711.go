// Package hx711 provides a driver for the HX711 24 bit, 2 channel, configurable ADC with serial output to measure small
// differential voltages. The device is handy for load cells but can be used to read all kind of Wheatstone bridges.
// Therefore the usage of the phrases "mass", "weight" or "load" are prevented in this driver - "value" is used instead.
//
// Datasheet: https://cdn.sparkfun.com/datasheets/Sensors/ForceFlex/hx711_english.pdf
package hx711

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"tinygo.org/x/drivers"
)

// Conversion from machine.Pin for TinyGo MCUs
//
// powerDownAndSckPin := machine.D0
// powerDownAndSckPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
// pinSetStateFunc := func(l bool) error {powerDownAndSckPin.Set(l); return nil}
type pinSetState func(newState bool) error

// Conversion from machine.Pin for TinyGo MCUs
//
// dataPin := machine.D1
// dataPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
// pinStateFunc := func() (bool, error) {return dataPin.Get(l), nil}
type pinState func() (bool, error)

type GainAndChannelCfg int

const (
	None    GainAndChannelCfg = 0 // channel is not used
	A128    GainAndChannelCfg = 1 // channel A, gain factor 128 - 25 pulses
	B32     GainAndChannelCfg = 2 // channel B, gain factor 32 - 26 pulses
	A64     GainAndChannelCfg = 3 // channel A, gain factor 64 - 27 pulses
	A128B32 GainAndChannelCfg = 4 // channel A@128 and channel B@32 - after first read 26 pulses, after second 25 pulses
	A64B32  GainAndChannelCfg = 5 // channel A@64 and channel B@32 - after first read 26 pulses, after second 27 pulses
)

const minResetDuration = 60 * time.Microsecond

type DeviceConfig struct {
	TickSleep     time.Duration // duration for H-part in the L-H-L pulse, see remarks in the default config below
	ResetDuration time.Duration // must not be smaller than 60 us
	SettlingTime  time.Duration
	ReadTimeout   time.Duration
	ReadTriesMax  uint8 // how often a check for ready state is done until the read timeout is reached
}

//nolint:gochecknoglobals,mnd // done by intention here
var DefaultConfig = DeviceConfig{
	// setting "tickSleep" to a value bigger than 0 is needed for fast MCUs, typically 1 us is needed
	// the HX711 works between 0.2 and 50 us pulse wide according the data sheet
	// e.g. for the nRF52840 setting this to 0 leads to a pulse wide of around 1 us, which is fine, but setting it
	// to 1 us leads to a pulse wide of 20 us, which also will work but slows down unnecessary
	TickSleep:     0,
	ResetDuration: 100 * time.Microsecond,
	SettlingTime:  400 * time.Millisecond, // for RATE=0 (10Hz), can be reduced to 50 ms for RATE=1 (80Hz)
	ReadTimeout:   2 * time.Second,
	ReadTriesMax:  100,
}

type readingConfig struct {
	gainAndChanAfterRead GainAndChannelCfg
	// offset and calibrationFactor will be used to adjust measures, we use this formulas:
	// * y = m*(x+n/m); n/m=offset; m=calibrationFactor
	// used for calibration:
	// * offset = y/m-x; offset=-x @ y=0
	// * m=y/(x+offset); y=setValue
	offset               int32
	calibrationFactorNum int32 // numerator, always != 0
	calibrationFactorDen int32 // denominator, always > 0
}

// Device contains attributes for reading the values of HX711.
type Device struct {
	// powerDownAndSckPinSetState is connected to the PD_SCK pin of the HX711, 25-27 pulses can be given
	// pulses are typically L-H-L for 1us (0.2..50us), when this pin is low, the chip is in normal operating mode
	// if clock stays more than 60us at high, the chip enters power down mode
	// 25 pulses: read 24 bits from the input with gain which was former selected, select A@128 for next read
	// 26 pulses: read 24 bits from the input with gain which was former selected, select B@32 for next read
	// 27 pulses: read 24 bits from the input with gain which was former selected, select A@64 for next read
	// note: after reset the input A with gain 128 is selected for next read (the first read is maybe not what you need)
	powerDownAndSckPinSetState pinSetState
	// dataPin is connected to the data output pin of the HX711, the 24 bits of data will be shifted out with MSB first
	// if the pin is high, the chip is not ready for data, e.g. after the 25th pulse or in power down mode
	// the pin should be configured as pull up or with an external pull up resistor - means not ready by default
	dataPinState pinState
	// device configuration options
	devCfg DeviceConfig
	// configuration of readings
	waitDuration     time.Duration
	firstReadingCfg  readingConfig
	secondReadingCfg readingConfig
	// synchronization
	mu *sync.Mutex
	// last stored value
	firstRawValue  int32 // can be the value from channel A or B (if only B was read)
	secondRawValue int32
}

// New returns a device for reading differential voltages with 2 inputs (A, B). The gain of input A can be chosen
// between 128 (default) and 64 - the gain of input B is always 32.
// The reading can be chosen between:
// * A@128 only
// * A@64 only
// * B@32 only
// * A@128 followed by B@32
// * A@64 followed by B@32
// The pins needs to be already configured at caller side
func New(powerDownAndSckPinSetState pinSetState, dataPinState pinState, gc GainAndChannelCfg) *Device {
	gc1, gc2 := gc.splitGainAndChannelConfig()
	rc1 := readingConfig{gainAndChanAfterRead: gc1, calibrationFactorNum: 1, calibrationFactorDen: 1}
	rc2 := readingConfig{gainAndChanAfterRead: gc2, calibrationFactorNum: 1, calibrationFactorDen: 1}

	d := Device{
		powerDownAndSckPinSetState: powerDownAndSckPinSetState,
		dataPinState:               dataPinState,
		devCfg:                     DefaultConfig,
		firstReadingCfg:            rc1,
		secondReadingCfg:           rc2,
		mu:                         &sync.Mutex{},
	}

	return &d
}

// Configure configures initially the driver
func (d *Device) Configure(cfg *DeviceConfig) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if cfg != nil {
		d.devCfg = *cfg
	}

	if d.devCfg.ResetDuration < minResetDuration {
		println("adapt reset duration to minimum required", minResetDuration.String())
		d.devCfg.ResetDuration = minResetDuration
	}

	if d.devCfg.ReadTriesMax < 1 {
		println("adapt maximum read tries to 1")
		d.devCfg.ReadTriesMax = 1
	}

	d.waitDuration = d.devCfg.ReadTimeout / time.Duration(d.devCfg.ReadTriesMax)

	return d.reset()
}

// Zero sets the offset for the reading. If the given flag is true, this is done for the second reading.
func (d *Device) Zero(secondReading bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	rc := &d.firstReadingCfg
	if secondReading {
		if d.secondReadingCfg.gainAndChanAfterRead == None {
			return errors.New("no zero possible, second reading is not configured")
		}

		rc = &d.secondReadingCfg
	}

	rc.set(0, 1, 1)

	v, v2, err := d.readChannelsWithTimout()
	if err != nil {
		return err
	}

	if secondReading {
		v = v2
	}

	rc.set(-v, 1, 1)

	return nil
}

// Calibrate calculates, after a measurement of the set value is done, a factor for linear scaling the values of the
// subsequent measurements. The unit of the given set value define the unit of the measurement result later. Before
// using this function, the offset value should be obtained by calling Zero() function with no load.
// If the given flag is true, this is done for the second reading.
func (d *Device) Calibrate(setValue int32, secondReading bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	rc := &d.firstReadingCfg
	if secondReading {
		if d.secondReadingCfg.gainAndChanAfterRead == None {
			return errors.New("no calibration possible, second reading is not configured")
		}

		rc = &d.secondReadingCfg
	}

	v, v2, err := d.readChannelsWithTimout()
	if err != nil {
		return err
	}

	if secondReading {
		v = v2
	}

	return rc.calculateAndSetCalibrationFactor(setValue, v)
}

// Update implements the drivers.Sensor interface
func (d *Device) Update(drivers.Measurement) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	v1, v2, err := d.readChannelsWithTimout()
	if err != nil {
		return err
	}

	d.firstRawValue = v1
	d.secondRawValue = v2

	return nil
}

// Values returns both scaled values from the last successful update
func (d *Device) Values() (int64, int64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.firstReadingCfg.scale(d.firstRawValue), d.secondReadingCfg.scale(d.secondRawValue)
}

// OffsetAndCalibrationFactor returns linear correction values, used for reading.
// If the given flag is true, this values are related to the second reading.
func (d *Device) OffsetAndCalibrationFactor(secondReading bool) (int32, float32) {
	d.mu.Lock()
	defer d.mu.Unlock()

	rc := d.firstReadingCfg
	if secondReading {
		rc = d.secondReadingCfg
	}

	o, cNum, cDen := rc.get()

	return o, float32(cNum) / float32(cDen)
}

// SetOffsetAndCalibrationFactor sets linear correction values, used for reading.
// If the given flag is true, this values are related to the second reading.
func (d *Device) SetOffsetAndCalibrationFactor(offset int32, calibrationFactor float32, secondReading bool) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	cfNum, cfDen, err := drivers.Float32Fractions(calibrationFactor)
	if err != nil {
		return err
	}

	rc := &d.firstReadingCfg
	if secondReading {
		rc = &d.secondReadingCfg
	}

	rc.set(offset, cfNum, cfDen)

	return nil
}

func (d *Device) reset() error {
	// set clock pin to high for >60us to enter power down mode (reset)
	if err := d.powerDownAndSckPinSetState(true); err != nil {
		return err
	}
	time.Sleep(d.devCfg.ResetDuration)
	// set the clock pin back to low to enter normal operating mode
	if err := d.powerDownAndSckPinSetState(false); err != nil {
		return err
	}
	time.Sleep(d.devCfg.SettlingTime)
	// make a first read, to apply the channel and gain configuration for the subsequent reads
	_, _, err := d.readChannelsWithTimout()

	return err
}

// readChannelsWithTimout performs a single read of both channels. If only the first reading is configured, only one
// reading is performed for the configured channel and the second return value will be zero.
func (d *Device) readChannelsWithTimout() (int32, int32, error) {
	v1, err := d.readWithTimout(d.firstReadingCfg.gainAndChanAfterRead)
	if err != nil || d.secondReadingCfg.gainAndChanAfterRead == None {
		return v1, 0, err
	}

	v2, err := d.readWithTimout(d.secondReadingCfg.gainAndChanAfterRead)

	return v1, v2, err
}

// readWithTimout waits for the device to be ready, reads the configured count of bits from the configured channel and
// returns it converted to an integer value. If the device is not ready in time, an error will be returned.
func (d *Device) readWithTimout(gc GainAndChannelCfg) (int32, error) {
	retries := d.devCfg.ReadTriesMax

	for retries > 0 {
		busy, err := d.dataPinState()
		if err != nil {
			return 0, err
		}
		if !busy {
			return d.read(gc)
		}
		time.Sleep(d.waitDuration)
		retries--
	}

	return 0, errors.New("timeout reached for HX711 on wait for ready state (" + strconv.Itoa(int(gc)) + ")")
}

// read reads the 24 bits serially with the configured count of ticks from the configured channel and returns the
// bits converted to an integer value.
func (d *Device) read(gc GainAndChannelCfg) (int32, error) {
	var value int32
	var bitSet bool
	var err error

	for i := 0; i < 24; i++ {
		if err := d.tick(); err != nil {
			return 0, err
		}
		value = value << 1
		bitSet, err = d.dataPinState()
		if err != nil {
			return 0, err
		}
		if bitSet {
			value = value | 1
		}
	}

	// write gain and channel
	for i := 0; i < int(gc); i++ {
		if err := d.tick(); err != nil {
			return 0, err
		}
	}

	//nolint:mnd // ok here
	value = (value << 8) >> 8 // ensure leading ones for negative value

	return value, nil
}

// tick creates a (L-)H-L pulse on the clock pin. For fast devices a configurable sleep is used.
func (d *Device) tick() error {
	if err := d.powerDownAndSckPinSetState(true); err != nil {
		return err
	}
	time.Sleep(d.devCfg.TickSleep)
	if err := d.powerDownAndSckPinSetState(false); err != nil {
		return err
	}
	time.Sleep(d.devCfg.TickSleep)

	return nil
}

// set sets a new offset and calibration factor to the configuration for reading
func (rc *readingConfig) set(o, cNum, cDen int32) {
	rc.offset = o
	rc.calibrationFactorNum = cNum
	rc.calibrationFactorDen = cDen
}

// get gets the configured offset and calibration factor for reading
func (rc *readingConfig) get() (int32, int32, int32) {
	return rc.offset, rc.calibrationFactorNum, rc.calibrationFactorDen
}

// calculateAndSetCalibrationFactor calculates the calibration factor from the difference between given set value and
// current value from a measurement
func (rc *readingConfig) calculateAndSetCalibrationFactor(setValue, currentRawValue int32) error {
	if setValue == 0 {
		return errors.New("set value needs to be <> 0")
	}

	if currentRawValue == 0 {
		return errors.New("read value is exactly 0")
	}

	// calibration factor can never become zero for 24 bit measure: 1/(8388607+2147483647) = ~4.6e-10
	rc.calibrationFactorNum = setValue
	rc.calibrationFactorDen = currentRawValue + rc.offset

	return nil
}

// scale adjust the given measurement value by the offset and calibration factor
func (rc *readingConfig) scale(v int32) int64 {
	return int64(v+rc.offset) * int64(rc.calibrationFactorNum) / int64(rc.calibrationFactorDen)
}

// ensure limit, splits and returns the config values to first and second reading
func (gc GainAndChannelCfg) splitGainAndChannelConfig() (GainAndChannelCfg, GainAndChannelCfg) {
	if gc < A128 {
		gc = A128
	}

	if gc > A64B32 {
		gc = A64B32
	}

	gc1 := gc
	gc2 := None
	//nolint:exhaustive // ok here
	switch gc {
	case A128B32:
		gc1 = B32
		gc2 = A128
	case A64B32:
		gc1 = B32
		gc2 = A64
	}

	return gc1, gc2
}
