// Package ltr390 provides a driver for the LTR-390UV ambient light and UV sensor by Lite-On.
//
// Datasheet: https://optoelectronics.liteon.com/upload/download/DS86-2015-0004/LTR-390UV_Final_%20DS_V1%201.pdf
package ltr390 // import "tinygo.org/x/drivers/ltr390"

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

var (
	ErrTimeout = errors.New("ltr390: timeout waiting for data")
)

// Sensitivity is 2300 counts per UVI at gain 18 and 20 bit, so UVI*1000 = UVS*18000/(2300*gain)*2^(20-bits). Datasheet section 4.5.
// 18000/2300 = 180/23, so this holds 180/gain for each gain and ReadUVIndex divides by 23.
var uvFactor = [...]uint32{180 / 1, 180 / 3, 180 / 6, 180 / 9, 180 / 18}

// Wakeup time from standby. Datasheet section 4.3.
const wakeupTime = 10 * time.Millisecond

var (
	gainFactor     = [...]uint32{1, 3, 6, 9, 18}
	resolutionBits = [...]uint8{20, 19, 18, 17, 16, 13}
	measRate       = [...]time.Duration{25, 50, 100, 200, 500, 1000, 2000}
)

// Device wraps an I2C connection to a LTR390 device.
type Device struct {
	bus        drivers.I2C
	Address    uint16
	gain       Gain
	resolution Resolution
	rate       MeasurementRate
	mode       Mode
	enabled    bool
	buf        [4]byte
}

// New creates a new LTR390 connection. The I2C bus must already be configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{
		bus:        bus,
		Address:    DefaultAddress,
		gain:       defaultGain,
		resolution: defaultResolution,
		rate:       defaultRate,
	}
}

// Connected returns whether a LTR390 has been found.
func (d *Device) Connected() bool {
	id, err := d.readReg(PartID)
	return err == nil && id>>4 == partNumberID
}

// Configure writes the configuration and enables the sensor in cfg.Mode.
func (d *Device) Configure(cfg Config) error {
	if cfg.Gain == 0 {
		cfg.Gain = defaultGain
	}
	if cfg.Resolution == 0 {
		cfg.Resolution = defaultResolution
	}
	if cfg.Rate == 0 {
		cfg.Rate = defaultRate
	}

	// The SW reset write is not ACKed and can hang some I2C peripherals, see
	// https://github.com/adafruit/Adafruit_LTR390/blob/master/Adafruit_LTR390.cpp
	if err := d.writeReg(MainCtrl, 0); err != nil {
		return err
	}
	d.enabled = false
	if err := d.writeReg(MeasRate, measRateValue(cfg.Resolution, cfg.Rate)); err != nil {
		return err
	}
	d.resolution = cfg.Resolution
	d.rate = cfg.Rate
	if err := d.writeReg(GainRange, byte(cfg.Gain-1)); err != nil {
		return err
	}
	d.gain = cfg.Gain
	if err := d.writeReg(IntCfg, intCfgALS); err != nil {
		return err
	}
	if err := d.writeReg(IntPst, 0); err != nil {
		return err
	}
	if err := d.SetThresholds(0, dataMask); err != nil {
		return err
	}
	return d.start(cfg.Mode)
}

// Enable puts the sensor in active mode.
func (d *Device) Enable() error {
	return d.start(d.mode)
}

// Disable puts the sensor in standby mode. The next read enables it again.
func (d *Device) Disable() error {
	ctrl := byte(0)
	if d.mode == ModeUVS {
		ctrl = ctrlUVSMode
	}
	d.enabled = false
	return d.writeReg(MainCtrl, ctrl)
}

// SetGain sets the analog gain of both channels.
func (d *Device) SetGain(gain Gain) error {
	if err := d.writeReg(GainRange, byte(gain-1)); err != nil {
		return err
	}
	d.gain = gain
	return d.restart()
}

// SetResolution sets the ADC resolution and conversion time.
func (d *Device) SetResolution(res Resolution) error {
	if err := d.writeReg(MeasRate, measRateValue(res, d.rate)); err != nil {
		return err
	}
	d.resolution = res
	return d.restart()
}

// SetMeasurementRate sets the time between measurements.
func (d *Device) SetMeasurementRate(rate MeasurementRate) error {
	if err := d.writeReg(MeasRate, measRateValue(d.resolution, rate)); err != nil {
		return err
	}
	d.rate = rate
	return d.restart()
}

// SetThresholds sets the lower and upper interrupt thresholds in raw counts.
func (d *Device) SetThresholds(low, high uint32) error {
	if err := d.write20bits(ThresLow, low); err != nil {
		return err
	}
	return d.write20bits(ThresUp, high)
}

// ConfigureInterrupt enables or disables the INT pin. It fires when the data of
// source is outside the thresholds for persistence+1 consecutive measurements.
func (d *Device) ConfigureInterrupt(enable bool, source Mode, persistence uint8) error {
	if persistence > 15 {
		persistence = 15
	}
	cfg := byte(intCfgALS)
	if source == ModeUVS {
		cfg = intCfgUVS
	}
	if enable {
		cfg |= intCfgEnable
	}
	if err := d.writeReg(IntCfg, cfg); err != nil {
		return err
	}
	return d.writeReg(IntPst, persistence<<4)
}

// DataReady returns whether there is new data. Reading the status clears the
// data ready and interrupt flags. Datasheet section 6.5.
func (d *Device) DataReady() (bool, error) {
	status, err := d.readReg(MainStatus)
	return status&statusDataReady != 0, err
}

// ReadRawALS switches to ALS mode if needed, waits for new data and returns the raw count.
func (d *Device) ReadRawALS() (uint32, error) {
	return d.read(ModeALS, ALSData)
}

// ReadRawUVS switches to UVS mode if needed, waits for new data and returns the raw count.
func (d *Device) ReadRawUVS() (uint32, error) {
	return d.read(ModeUVS, UVSData)
}

// ReadIlluminance returns the ambient light in milli lux.
func (d *Device) ReadIlluminance() (uint32, error) {
	als, err := d.ReadRawALS()
	if err != nil {
		return 0, err
	}
	// Lux = 0.6 * ALS / (GAIN * INT) with INT = 2^(bits-18). Datasheet section 7.1.
	// 13 bit is not in the datasheet table and follows the same scaling.
	mlux := als * 600
	den := gainFactor[d.gain-1]
	if bits := resolutionBits[d.resolution-1]; bits > 18 {
		den <<= bits - 18
	} else {
		mlux <<= 18 - bits
	}
	return mlux / den, nil
}

// ReadUVIndex returns the UV index multiplied by 1000.
func (d *Device) ReadUVIndex() (uint32, error) {
	uvs, err := d.ReadRawUVS()
	if err != nil {
		return 0, err
	}
	// UVI = UVS / sensitivity, scaled from gain 18 and 20 bit. Datasheet section 7.2.
	return (uvs * uvFactor[d.gain-1] << (20 - resolutionBits[d.resolution-1])) / 23, nil
}

func (d *Device) read(mode Mode, reg uint8) (uint32, error) {
	if !d.enabled || d.mode != mode {
		if err := d.start(mode); err != nil {
			return 0, err
		}
	}
	if err := d.waitDataReady(); err != nil {
		return 0, err
	}
	d.buf[0] = reg
	if err := d.bus.Tx(d.Address, d.buf[:1], d.buf[1:4]); err != nil {
		return 0, err
	}
	return (uint32(d.buf[1]) | uint32(d.buf[2])<<8 | uint32(d.buf[3])<<16) & dataMask, nil
}

// start writes MAIN_CTRL, which starts a new measurement (datasheet section 6.1),
// then reads MAIN_STATUS to clear a stale data ready flag.
func (d *Device) start(mode Mode) error {
	ctrl := byte(ctrlEnable)
	if mode == ModeUVS {
		ctrl |= ctrlUVSMode
	}
	if err := d.writeReg(MainCtrl, ctrl); err != nil {
		return err
	}
	d.mode = mode
	d.enabled = true
	_, err := d.readReg(MainStatus)
	return err
}

// restart discards a measurement done with old settings.
func (d *Device) restart() error {
	if !d.enabled {
		return nil
	}
	return d.start(d.mode)
}

func (d *Device) waitDataReady() error {
	timeout := 2*d.period() + wakeupTime
	start := time.Now()
	for {
		ready, err := d.DataReady()
		if err != nil {
			return err
		}
		if ready {
			return nil
		}
		if time.Since(start) > timeout {
			return ErrTimeout
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// period returns the time between measurements. Datasheet section 6.2.
func (d *Device) period() time.Duration {
	conversion := (400 * time.Millisecond) >> (d.resolution - 1)
	rate := measRate[d.rate-1] * time.Millisecond
	if conversion > rate {
		return conversion
	}
	return rate
}

func (d *Device) readReg(reg uint8) (byte, error) {
	d.buf[0] = reg
	err := d.bus.Tx(d.Address, d.buf[:1], d.buf[1:2])
	return d.buf[1], err
}

func (d *Device) writeReg(reg uint8, value byte) error {
	d.buf[0] = reg
	d.buf[1] = value
	return d.bus.Tx(d.Address, d.buf[:2], nil)
}

func (d *Device) write20bits(reg uint8, value uint32) error {
	value &= dataMask
	d.buf[0] = reg
	d.buf[1] = byte(value)
	d.buf[2] = byte(value >> 8)
	d.buf[3] = byte(value >> 16)
	return d.bus.Tx(d.Address, d.buf[:4], nil)
}

func measRateValue(res Resolution, rate MeasurementRate) byte {
	return byte(res-1)<<4 | byte(rate-1)
}
