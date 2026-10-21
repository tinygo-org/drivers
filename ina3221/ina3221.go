// Package ina3221 implements a driver for the TI INA3221 three channel
// shunt and bus voltage monitor.
//
// Datasheet: https://www.ti.com/lit/ds/symlink/ina3221.pdf
package ina3221 // import "tinygo.org/x/drivers/ina3221"

import (
	"errors"
	"math"

	"tinygo.org/x/drivers"
)

// Channel numbers match the datasheet.
const (
	Channel1 = 1
	Channel2 = 2
	Channel3 = 3
)

// DefaultShuntResistance is the shunt value on most breakout boards.
const DefaultShuntResistance = 100 // milliohms

var (
	ErrInvalidChannel     = errors.New("ina3221: invalid channel")
	ErrInvalidResistance  = errors.New("ina3221: invalid shunt resistance")
	ErrConfigMismatch     = errors.New("ina3221: config mismatch")
	ErrDeviceNotConnected = errors.New("ina3221: device not connected")
)

// Device wraps an I2C connection to an INA3221 device.
type Device struct {
	bus     drivers.I2C
	Address uint16
	config  Config
	shunt   [3]int32
	buf     [3]byte
}

// New creates a new INA3221 connection. The I2C bus must already be
// configured. It does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
		config:  DefaultConfig,
		shunt:   [3]int32{DefaultShuntResistance, DefaultShuntResistance, DefaultShuntResistance},
	}
}

// Configure checks the device IDs, then writes and verifies the configuration.
func (d *Device) Configure(cfg Config) error {
	if !d.Connected() {
		return ErrDeviceNotConnected
	}

	if err := d.WriteRegister(RegConfig, cfg.RegisterValue()); err != nil {
		return err
	}

	val, err := d.ReadRegister(RegConfig)
	if err != nil {
		return err
	}
	if val != cfg.RegisterValue() {
		return ErrConfigMismatch
	}

	d.config = cfg
	return nil
}

// Config returns the last configuration written by Configure.
func (d *Device) Config() Config {
	return d.config
}

// Reset sets all registers to their default values.
func (d *Device) Reset() error {
	if err := d.WriteRegister(RegConfig, configReset); err != nil {
		return err
	}
	d.config = DefaultConfig
	return nil
}

// Connected reports whether an INA3221 answers at Address.
func (d *Device) Connected() bool {
	manf, err := d.ReadRegister(RegManufacturerID)
	if err != nil || manf != ManufacturerID {
		return false
	}
	die, err := d.ReadRegister(RegDieID)
	return err == nil && die == DieID
}

// Trigger starts a conversion in a single-shot mode.
// It does nothing in a continuous mode.
func (d *Device) Trigger() error {
	if !d.config.Mode.SingleShot() {
		return nil
	}
	return d.WriteRegister(RegConfig, d.config.RegisterValue())
}

// ConversionReady reports whether all enabled conversions are complete.
// Reading it clears the flag.
func (d *Device) ConversionReady() (bool, error) {
	val, err := d.ReadRegister(RegMaskEnable)
	if err != nil {
		return false, err
	}
	return val&maskConversionReady != 0, nil
}

// SetShuntResistance sets the shunt resistor value in milliohms
// used by Current and Power for the given channel.
func (d *Device) SetShuntResistance(channel int, milliohms int32) error {
	if !validChannel(channel) {
		return ErrInvalidChannel
	}
	if milliohms <= 0 {
		return ErrInvalidResistance
	}
	d.shunt[channel-1] = milliohms
	return nil
}

// BusVoltage returns the bus voltage of a channel in µV.
func (d *Device) BusVoltage(channel int) (int32, error) {
	if !validChannel(channel) {
		return 0, ErrInvalidChannel
	}
	val, err := d.ReadRegister(RegBusVoltage1 + uint8(channel-1)*2)
	if err != nil {
		return 0, err
	}
	// The LSB is 8 mV, see the bus voltage registers in datasheet section 8.6.
	return int32(int16(val)>>3) * 8000, nil
}

// ShuntVoltage returns the shunt voltage of a channel in µV.
func (d *Device) ShuntVoltage(channel int) (int32, error) {
	if !validChannel(channel) {
		return 0, ErrInvalidChannel
	}
	val, err := d.ReadRegister(RegShuntVoltage1 + uint8(channel-1)*2)
	if err != nil {
		return 0, err
	}
	// The LSB is 40 µV, see the shunt voltage registers in datasheet section 8.6.
	return int32(int16(val)>>3) * 40, nil
}

// Current returns the current of a channel in µA.
func (d *Device) Current(channel int) (int32, error) {
	shunt, err := d.ShuntVoltage(channel)
	if err != nil {
		return 0, err
	}
	return shunt * 1000 / d.shunt[channel-1], nil
}

// Power returns the power of a channel in µW.
func (d *Device) Power(channel int) (int32, error) {
	_, _, _, power, err := d.Measurements(channel)
	return power, err
}

// Measurements returns the bus voltage in µV, shunt voltage in µV,
// current in µA, and power in µW of a channel.
func (d *Device) Measurements(channel int) (bus, shunt, current, power int32, err error) {
	if bus, err = d.BusVoltage(channel); err != nil {
		return
	}
	if shunt, err = d.ShuntVoltage(channel); err != nil {
		return
	}
	current = shunt * 1000 / d.shunt[channel-1]
	p := int64(bus) * int64(current) / 1000000
	switch {
	case p > math.MaxInt32:
		power = math.MaxInt32
	case p < math.MinInt32:
		power = math.MinInt32
	default:
		power = int32(p)
	}
	return
}

// ReadRegister reads a 16-bit register.
func (d *Device) ReadRegister(reg uint8) (uint16, error) {
	d.buf[0] = reg
	if err := d.bus.Tx(d.Address, d.buf[:1], d.buf[1:3]); err != nil {
		return 0, err
	}
	return uint16(d.buf[1])<<8 | uint16(d.buf[2]), nil
}

// WriteRegister writes a 16-bit register.
func (d *Device) WriteRegister(reg uint8, val uint16) error {
	d.buf[0] = reg
	d.buf[1] = byte(val >> 8)
	d.buf[2] = byte(val)
	return d.bus.Tx(d.Address, d.buf[:3], nil)
}

func validChannel(channel int) bool {
	return channel >= Channel1 && channel <= Channel3
}
