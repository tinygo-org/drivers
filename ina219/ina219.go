package ina219

import (
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// An INA219 device.
type Device struct {
	bus     drivers.I2C
	Address uint16
	config  Config
}

// Create a new INA219 device with the default configuration
// and the given I2C bus at the default address.
//
// Set Address after New to change the address.
//
// Call Configure after New to write the configuration to the
// device. If you don't call Configure, the device may have a
// different configuration and the power divider and current
// multiplier are probably wrong.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
		config:  Config32V2A,
	}
}

// Set the configuration for the device. This only changes the
// configuration in memory, not on the device. Call Configure
// to write the configuration to the device.
func (d *Device) SetConfig(config Config) {
	d.config = config
}

// Write the current configuration to the device.
func (d *Device) Configure() (err error) {
	if err = d.WriteRegister(
		RegConfig,
		d.config.RegisterValue(),
	); err != nil {
		return
	}

	if err = d.WriteRegister(
		RegCalibration,
		d.config.Calibration.RegisterValue(),
	); err != nil {
		return
	}

	var readConfig Config
	// make sure the configuration is read back correctly
	if readConfig, err = d.ReadConfig(); err != nil {
		return
	} else if readConfig.RegisterValue() != d.config.RegisterValue() {
		err = ErrConfigMismatch{}
	} else if readConfig.Calibration.RegisterValue() != d.config.Calibration.RegisterValue() {
		err = ErrConfigMismatch{}
	}

	return
}

// Trigger a conversion. This is only necessary if the device is in
// trigger mode. In continuous mode (the default), the device will
// automatically trigger conversions and this has no effect. See
// config.go.
//
// Triggering a conversion or reading the "power" register resets
// the conversion ready bit.
func (d *Device) Trigger() (err error) {
	// Only trigger if the mode is one of the triggered modes.
	if ModeTriggered(d.config.Mode) {
		err = d.WriteRegister(RegConfig, d.config.RegisterValue())
	}
	return
}

// Measurements reads the bus voltage, shunt voltage, current, and power
// from the device.
func (d *Device) Measurements() (
	busVoltage int16,
	shuntVoltage int16,
	current float32,
	power float32,
	err error,
) {
	// Attempt to read bus voltage first, so we can check for overflow
	// or conversion not ready.
	if busVoltage, err = d.BusVoltage(); err != nil {
		return
	}

	// Read the rest of the values, reading Power last, which resets
	// the conversion ready bit (relevant for triggered modes).
	if shuntVoltage, err = d.ShuntVoltage(); err != nil {
		return
	}

	if current, err = d.Current(); err != nil {
		return
	}

	if power, err = d.Power(); err != nil {
		return
	}

	return
}

// BusVoltage reads the "bus" voltage in millivolts.
//
// It returns an error if the value is invalid due to overflow
// or if the conversion is not ready yet. In a continuous mode
// there should always be a measurement available after the
// device is ready. See above notes on Trigger.
func (d *Device) BusVoltage() (voltage int16, err error) {
	val, err := d.ReadRegister(RegBusVoltage)
	if err != nil {
		return
	}

	// The overflow bit is set, so the values are invalid.
	if val&(1<<0) != 0 {
		err = ErrOverflow{}
		return
	}

	// The conversion is not ready yet.
	if ModeTriggered(d.config.Mode) && val&(1<<1) != 0 {
		err = ErrNotReady{}
		return
	}

	voltage = (int16(val) >> 3) * 4
	return
}

// ShuntVoltage reads the "shunt" voltage in 100ths of a millivolt.
func (d *Device) ShuntVoltage() (voltage int16, err error) {
	return d.ReadRegister(RegShuntVoltage)
}

// Current reads the current in milliamps.
func (d *Device) Current() (current float32, err error) {
	val, err := d.ReadRegister(RegCurrent)
	if err != nil {
		return
	}

	current = float32(val) / d.config.CurrentDivider
	return
}

// Power reads the power in milliwatts.
func (d *Device) Power() (power float32, err error) {
	val, err := d.ReadRegister(RegPower)
	if err != nil {
		return
	}

	power = float32(val) * d.config.PowerMultiplier
	return
}

// Read the configuration from the device.
func (d *Device) ReadConfig() (config Config, err error) {
	var cfg, cal int16
	if cfg, err = d.ReadRegister(RegConfig); err != nil {
		return
	}

	if cal, err = d.ReadRegister(RegCalibration); err != nil {
		return
	}

	config = NewConfig(cfg, cal)
	return
}

// Read a register from the device.
func (d *Device) ReadRegister(reg uint8) (val int16, err error) {
	buf := make([]byte, 2)

	err = legacy.ReadRegister(d.bus, uint8(d.Address), reg, buf)
	if err != nil {
		return
	}

	val = int16(buf[0])<<8 | int16(buf[1]&0xff)
	return
}

// Write to a register on the device.
func (d *Device) WriteRegister(reg uint8, val uint16) error {
	buf := []byte{byte(val >> 8), byte(val & 0xff)}
	return legacy.WriteRegister(d.bus, uint8(d.Address), reg, buf)
}
