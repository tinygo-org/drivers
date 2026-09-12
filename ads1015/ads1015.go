// Package ads1015 provides a driver for the ADS1015 4-channel 12-bit ADC
//
// This driver is based on https://github.com/RobTillaart/ADS1X15
// Datasheet: https://www.ti.com/lit/ds/symlink/ads1015.pdf
package ads1015 // import "tinygo.org/x/drivers/ads1015"

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

// The datasheet defines the LSB as FS/2048, but we divide by the maximum
// positive code (2047) so that the maximum ADC output maps to full scale.
const maxCode = 2047

const conversionTimeout = 100 * time.Millisecond

var (
	ErrInvalidChannel = errors.New("ads1015: invalid channel")
	ErrTimeout        = errors.New("ads1015: conversion timeout")
)

// Config contains the ADS1015 conversion settings.
type Config struct {
	// Gain selects the full-scale input range of the PGA.
	Gain Gain

	// Mode selects between continuous and single-shot conversions.
	Mode Mode

	// DataRate sets the output data rate.
	DataRate DataRate

	// ComparatorMode, ComparatorPolarity, ComparatorLatch, and
	// ComparatorQueue configure the ALERT/RDY comparator. Leave
	// ComparatorQueue at ComparatorQueueDisable (the default) to disable
	// the comparator.
	ComparatorMode     ComparatorMode
	ComparatorPolarity ComparatorPolarity
	ComparatorLatch    ComparatorLatch
	ComparatorQueue    ComparatorQueue

	conversionPollInterval time.Duration
}

// DefaultConfig selects the widest gain range (+/-6.144V), single-shot
// mode, 1600 SPS, and disables the comparator.
var DefaultConfig = Config{
	Gain:               Gain6144mV,
	Mode:               ModeSingle,
	DataRate:           DataRate1600SPS,
	ComparatorMode:     ComparatorModeTraditional,
	ComparatorPolarity: ComparatorPolarityActiveLow,
	ComparatorLatch:    ComparatorNonLatching,
	ComparatorQueue:    ComparatorQueueDisable,
}

// Device is an ADS1015 ADC connected over I2C.
type Device struct {
	bus     drivers.I2C
	Address uint16
	config  Config
}

// New returns a new ADS1015 driver using DefaultConfig.
// Set Address after New if the ADDR pin is different
func New(bus drivers.I2C) *Device {
	config := DefaultConfig
	config.conversionPollInterval = ConversionDuration(config.DataRate)

	return &Device{
		bus:     bus,
		Address: Address,
		config:  config,
	}
}

// Config returns the currently stored configuration.
func (d *Device) Config() Config {
	return d.config
}

// Configure stores config and writes it to the device
func (d *Device) Configure(config Config) error {
	d.config = config
	d.config.conversionPollInterval = ConversionDuration(d.config.DataRate)
	return d.startConversion(muxSingleEnded[0])
}

// Connected reports whether an ADS1015 responds on the I2C bus.
func (d *Device) Connected() bool {
	_, err := d.readRegister(regConfig)
	return err == nil
}

// SetThresholdLow sets the low threshold used by the comparator.
func (d *Device) SetThresholdLow(v uint16) error {
	return d.writeRegister(regLowThreshold, v)
}

// ThresholdLow returns the low threshold used by the comparator.
func (d *Device) ThresholdLow() (int16, error) {
	v, err := d.readRegister(regLowThreshold)
	return int16(v), err
}

// SetThresholdHigh sets the high threshold used by the comparator.
func (d *Device) SetThresholdHigh(v uint16) error {
	return d.writeRegister(regHighThreshold, v)
}

// ThresholdHigh returns the high threshold used by the comparator.
func (d *Device) ThresholdHigh() (int16, error) {
	v, err := d.readRegister(regHighThreshold)
	return int16(v), err
}

// EnableConversionReadyPin repurposes the ALERT/RDY pin: instead of acting
// as a threshold comparator, it pulses once per completed conversion. This
// lets external hardware (e.g. an MCU interrupt) detect a finished
// conversion without polling Ready() over I2C.
//
// It does so by writing the high and low threshold registers to the
// special values documented in the datasheet (high threshold's MSB set,
// low threshold's MSB clear), and, if needed, moving ComparatorQueue off
// ComparatorQueueDisable so the pin stays active; the new ComparatorQueue
// takes effect starting with the next conversion.
func (d *Device) EnableConversionReadyPin() error {
	if d.config.ComparatorQueue == ComparatorQueueDisable {
		d.config.ComparatorQueue = ComparatorQueueAfter1Conv
	}
	if err := d.SetThresholdHigh(conversionReadyHiThresh); err != nil {
		return err
	}
	return d.SetThresholdLow(conversionReadyLoThresh)
}

// Read performs a conversion using the given mux setting and returns the
// raw signed 12-bit result. Use ReadADC or ReadADCDifferentialXX for the
// common cases.
func (d *Device) Read(mux Mux) (int16, error) {
	if err := d.startConversion(mux); err != nil {
		return 0, err
	}

	if d.config.Mode == ModeSingle {
		// Allow the device to clear the OS bit after starting the conversion.
		time.Sleep(d.config.conversionPollInterval)

		start := time.Now()

		for {
			ready, err := d.Ready()
			if err != nil {
				return 0, err
			}
			if ready {
				break
			}
			if time.Since(start) > conversionTimeout {
				return 0, ErrTimeout
			}
			time.Sleep(d.config.conversionPollInterval)
		}
	} else {
		// In continuous mode, give the device time to complete a
		// conversion at the new mux setting; otherwise a stale value left
		// over from the previous mux setting would be returned.
		time.Sleep(d.config.conversionPollInterval)
	}

	return d.Value()
}

// ReadADC performs a conversion on the given single-ended channel (0-3) and
// returns the raw signed 12-bit result.
func (d *Device) ReadADC(channel uint8) (int16, error) {
	if channel > 3 {
		return 0, ErrInvalidChannel
	}
	return d.Read(muxSingleEnded[channel])
}

// ReadVoltage performs a conversion on the given single-ended channel (0-3)
// and returns the result in millivolts, using the currently configured gain.
func (d *Device) ReadVoltage(channel uint8) (int32, error) {
	raw, err := d.ReadADC(channel)
	if err != nil {
		return 0, err
	}
	return d.ToVoltage(raw), nil
}

// ReadADCDifferential01 performs a differential conversion between AIN0 and
// AIN1 and returns the raw signed 12-bit result.
func (d *Device) ReadADCDifferential01() (int16, error) {
	return d.Read(MuxDiff01)
}

// ReadADCDifferential03 performs a differential conversion between AIN0 and
// AIN3 and returns the raw signed 12-bit result.
func (d *Device) ReadADCDifferential03() (int16, error) {
	return d.Read(MuxDiff03)
}

// ReadADCDifferential13 performs a differential conversion between AIN1 and
// AIN3 and returns the raw signed 12-bit result.
func (d *Device) ReadADCDifferential13() (int16, error) {
	return d.Read(MuxDiff13)
}

// ReadADCDifferential23 performs a differential conversion between AIN2 and
// AIN3 and returns the raw signed 12-bit result.
func (d *Device) ReadADCDifferential23() (int16, error) {
	return d.Read(MuxDiff23)
}

// ToVoltage converts a raw reading, such as one returned by ReadADC, to
// milli volts using the currently configured gain.
func (d *Device) ToVoltage(raw int16) int32 {
	return (int32(raw) * d.config.Gain.FullScaleVoltage()) / maxCode
}

// Request starts a conversion using the given mux setting without waiting
// for it to complete. Use Ready and Value to retrieve the result once it is
// available; Read does both in one call.
func (d *Device) Request(mux Mux) error {
	return d.startConversion(mux)
}

// RequestADC starts a conversion on the given single-ended channel (0-3)
// without waiting for it to complete.
func (d *Device) RequestADC(channel uint8) error {
	if channel > 3 {
		return ErrInvalidChannel
	}
	return d.Request(muxSingleEnded[channel])
}

// Ready reports whether the most recently requested conversion has finished.
// In single-shot mode, it reports whether the conversion is complete.
// In continuous mode, it reports whether the current conversion has finished;
// it may therefore return false while a conversion is in progress.
func (d *Device) Ready() (bool, error) {
	config, err := d.readRegister(regConfig)
	if err != nil {
		return false, err
	}
	return config&configOSReady != 0, nil
}

// Value reads the most recent conversion result from the conversion
// register, without starting a new conversion.
func (d *Device) Value() (int16, error) {
	raw, err := d.readRegister(regConversion)
	if err != nil {
		return 0, err
	}
	// The 12-bit result is left-justified in the 16-bit register.
	return int16(raw) >> 4, nil
}

// startConversion writes the configuration register to (re)start a
// conversion using the given mux setting and the stored configuration.
func (d *Device) startConversion(mux Mux) error {
	config := configOSStart |
		uint16(mux) |
		uint16(d.config.Gain) |
		uint16(d.config.Mode) |
		uint16(d.config.DataRate) |
		uint16(d.config.ComparatorMode) |
		uint16(d.config.ComparatorPolarity) |
		uint16(d.config.ComparatorLatch) |
		uint16(d.config.ComparatorQueue)
	return d.writeRegister(regConfig, config)
}

func (d *Device) readRegister(reg uint8) (uint16, error) {
	buf := make([]byte, 2)
	if err := d.bus.Tx(d.Address, []byte{reg}, buf); err != nil {
		return 0, err
	}
	return uint16(buf[0])<<8 | uint16(buf[1]), nil
}

func (d *Device) writeRegister(reg uint8, value uint16) error {
	return d.bus.Tx(d.Address, []byte{reg, byte(value >> 8), byte(value)}, nil)
}

// ConversionDuration returns the nominal conversion period derived from the
// data rate specified in the ADS1015 datasheet.
// The data rates are specified in Table 8-3; the conversion periods below
// are calculated as 1 / SPS and rounded to the nearest microsecond.
func ConversionDuration(dataRate DataRate) time.Duration {
	switch dataRate {
	case DataRate128SPS:
		return 7813 * time.Microsecond
	case DataRate250SPS:
		return 4000 * time.Microsecond
	case DataRate490SPS:
		return 2041 * time.Microsecond
	case DataRate920SPS:
		return 1087 * time.Microsecond
	case DataRate1600SPS:
		return 625 * time.Microsecond
	case DataRate2400SPS:
		return 417 * time.Microsecond
	case DataRate3300SPS:
		return 303 * time.Microsecond
	default:
		return 8 * time.Millisecond
	}
}
