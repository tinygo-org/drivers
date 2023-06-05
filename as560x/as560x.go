// Package as560x implements drivers for the ams AS5600/AS5601 on-axis magnetic rotary position sensors
//
// Product Pages:
//	AS5600: https://ams.com/as5600
//	AS5601: https://ams.com/as5601
//
// Datasheets:
//	AS5600: https://ams.com/documents/20143/36005/AS5600_DS000365_5-00.pdf
//	AS5601: https://ams.com/documents/20143/36005/AS5601_DS000395_3-00.pdf
//

package as560x // import tinygo.org/x/drivers/ams560x

import (
	"errors"

	"tinygo.org/x/drivers"
)

// Config holds the configuration for the AMS AS560x sensor devices.
type Config struct {
	// Address is the I2C address of the AS560x device. If left zero this will default to 0x36
	Address uint8
}

// MagnetStrength is an enum to indicate the magnetic field strength detected by the AS560x sensors.
type MagnetStrength int

const (
	// MagnetTooWeak indicates that the magnet strength is too weak (AGC maximum gain overflow) - move it closer
	MagnetTooWeak MagnetStrength = iota - 1
	// MagnetOk indicates that the magnet strength is about right.
	MagnetOk
	// MagnetTooStrong indicates that the magnet strength is too strong (AGC minimum gain overflow) - move it further away
	MagnetTooStrong
)

// AngleUnit is an enum to allow the use of different units when reading/writing angles from the AS560x sensors.
type AngleUnit int

const (
	// ANGLE_NATIVE uses the device's native angle measurement. i.e. 12-bit integer, 0 <= angle <= 0xfff (4095)
	ANGLE_NATIVE AngleUnit = iota
	// ANGLE_DEGREES_INT measures angles in degrees using integer arithmetic for speed. i.e. 0 <= angle < 360
	ANGLE_DEGREES_INT
	// ANGLE_DEGREES_FLOAT measures angles in degrees using floating point (slower). i.e. 0.0 <= angle < 360.0
	ANGLE_DEGREES_FLOAT
	// ANGLE_RADIANS measures angles in radians using floating point (slower). i.e. 0.0 <= angle < 2 * PI
	ANGLE_RADIANS
)

const (
	// NATIVE_ANGLE_MAX is the maximum valid value for a native angle for a AS560x device
	NATIVE_ANGLE_MAX = (1 << 12) - 1 + iota
	// NATIVE_ANGLE_RANGE is the number of unique values for native angles for a AS560x device
	NATIVE_ANGLE_RANGE
)

var (
	errRegisterNotFound = errors.New("Register not found")
	errMaxBurnAngle     = errors.New("Max BURN_ANGLE limit reached")
)

// BaseDevice handles the common behaviour between AS5600 & AS5601 devices
type BaseDevice struct {
	bus       drivers.I2C
	address   uint8
	registers map[uint8]*i2cRegister
	maxAngle  uint16
}

// newBaseDevice creates a new base device given an I2C bus.
func newBaseDevice(bus drivers.I2C) BaseDevice {
	// Add all 'base' registers, common to both AS5600 & AS5601
	conf := newI2CRegister(CONF, 0, 0x3fff, 2, reg_read|reg_write|reg_program)
	status := newI2CRegister(STATUS, 0, 0xff, 1, reg_read)
	regs := map[uint8]*i2cRegister{
		ZPOS:      newI2CRegister(ZPOS, 0, 0xfff, 2, reg_read|reg_write|reg_program),
		CONF:      conf,
		RAW_ANGLE: newI2CRegister(RAW_ANGLE, 0, 0xfff, 2, reg_read),
		ANGLE:     newI2CRegister(ANGLE, 0, 0xfff, 2, reg_read),
		STATUS:    status,
		AGC:       newI2CRegister(AGC, 0, 0xff, 1, reg_read),
		MAGNITUDE: newI2CRegister(MAGNITUDE, 0, 0xfff, 2, reg_read),
		BURN:      newI2CRegister(BURN, 0, 0xff, 1, reg_write),
		// Add common 'virtual registers' These are bitfields within the common registers above
		// A virtual register provides a convenient way to access the fields of a registers
		// by handling all of the necessary bitfield shifting and masking operations
		WD:   newVirtualRegister(conf, 13, 0b1),
		FTH:  newVirtualRegister(conf, 10, 0b111),
		SF:   newVirtualRegister(conf, 8, 0b11),
		HYST: newVirtualRegister(conf, 2, 0b11),
		PM:   newVirtualRegister(conf, 0, 0b11),
		MD:   newVirtualRegister(status, 5, 0b1),
		ML:   newVirtualRegister(status, 4, 0b1),
		MH:   newVirtualRegister(status, 3, 0b1),
	}
	return BaseDevice{bus, DefaultAddress, regs, NATIVE_ANGLE_RANGE}
}

// Configure sets up the AMS AS560x sensor device with the given configuration.
func (d *BaseDevice) Configure(cfg Config) {
	if cfg.Address == 0 {
		cfg.Address = DefaultAddress
	}
	d.address = cfg.Address
}

// ReadRegister reads the value for the given register from the AS560x device via I2C
func (d *BaseDevice) ReadRegister(address uint8) (uint16, error) {
	reg, ok := d.registers[address]
	if !ok {
		return 0, errRegisterNotFound
	}
	return reg.read(d.bus, d.address)
}

// WriteRegister writes the given value for the given register to the AS560x device via I2C
func (d *BaseDevice) WriteRegister(address uint8, value uint16) error {
	reg, ok := d.registers[address]
	if !ok {
		return errRegisterNotFound
	}
	return reg.write(d.bus, d.address, value)
}

// GetZeroPosition returns the 'zero position' (ZPOS) in various units
func (d *BaseDevice) GetZeroPosition(units AngleUnit) (uint16, float32, error) {
	zpos, err := d.ReadRegister(ZPOS)
	if nil != err {
		return 0, 0.0, err
	}
	// Convert to requested units
	i, f := convertFromNativeAngle(zpos, NATIVE_ANGLE_RANGE, units)
	return i, f, nil
}

// SetZeroPosition sets the 'zero position' (ZPOS) in various units
func (d *BaseDevice) SetZeroPosition(zpos float32, units AngleUnit) error {
	return d.WriteRegister(ZPOS, convertToNativeAngle(zpos, units))
}

// RawAngle reads the (unscaled & unadjusted) RAW_ANGLE register in various units
func (d *BaseDevice) RawAngle(units AngleUnit) (uint16, float32, error) {
	angle, err := d.ReadRegister(RAW_ANGLE)
	if nil != err {
		return 0, 0.0, err
	}
	// Convert to requested units
	i, f := convertFromNativeAngle(angle, NATIVE_ANGLE_RANGE, units)
	return i, f, nil
}

// Angle reads the (scaled & adjusted) ANGLE register in various units
func (d *BaseDevice) Angle(units AngleUnit) (uint16, float32, error) {
	// ZPOS enables setting the 'zero position' of the device to any RAW_ANGLE value
	// ANGLE is RAW_ANGLE adjusted relative to ZPOS.
	angle, err := d.ReadRegister(ANGLE)
	if nil != err {
		return 0, 0.0, err
	}
	// Convert to requested units
	i, f := convertFromNativeAngle(angle, d.maxAngle, units)
	return i, f, nil
}

// MagnetStatus reads the STATUS register and reports magnet position characteristics
func (d *BaseDevice) MagnetStatus() (detected bool, strength MagnetStrength, err error) {
	status, err := d.ReadRegister(STATUS)
	if nil != err {
		return false, MagnetOk, err
	}
	detected = (status & STATUS_MD) != 0
	strength = MagnetOk
	if (status & STATUS_ML) != 0 {
		strength = MagnetTooWeak
	} else if (status & STATUS_MH) != 0 {
		strength = MagnetTooStrong
	}
	return
}

// Burn is a convenience method to program the device permanently by writing to the BURN register (limited number of times use!)
func (d *BaseDevice) Burn(burnCmd BURN_CMD) error {
	if BURN_ANGLE == burnCmd {
		// BURN_ANGLE can only be executed up to 3 times.
		// We can check this in advance by reading ZMCO before writing to the BURN register.
		numBurns, err := d.ReadRegister(ZMCO)
		if nil != err {
			return err
		}
		if numBurns >= BURN_ANGLE_COUNT_MAX {
			// We're outta BURNs :(
			return errMaxBurnAngle
		}
	}
	return d.WriteRegister(BURN, uint16(burnCmd))
}
