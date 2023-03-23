//	Product: https://ams.com/as5600
//	Datasheet: https://ams.com/documents/20143/36005/AS5600_DS000365_5-00.pdf

package as560x // import tinygo.org/x/drivers/ams560x

import (
	"time"

	"tinygo.org/x/drivers"
)

// AS5600 includes MPOS & MANG in addition to ZPOS to set a 'narrower angle range'
// ZPOS enables setting the 'zero position' of the device to any RAW_ANGLE value.
// MPOS ('max position') & MANG 'max angle' enable a 'partial range' on the AS5600.
// The value in ANGLE is scaled & adjusted by the device according to ZPOS and MPOS/MANG.
// The entire 12-bit range is 'compressed' into the RAW_ANGLE range of ZPOS->MPOS
// (or ZPOS->ZPOS+MANG) thus enabling a higher resolution for a partial range.
// if ZPOS > MPOS (or ZPOS + MANG > 4095) i.e. the incremental range 'crosses zero'
// then the device will automatically compensate for the correct range.
// For RAW_ANGLE values outside of the partial range, ANGLE will be 'capped' at either
// 0 or 4095, depending on 'which end of the partial range is closer.'

// AS5600Device represents an ams AS5600 device driver accessed over I2C
type AS5600Device struct {
	// promote BaseDevice
	BaseDevice
}

// NewAS5600 creates a new AS5600Device given an I2C bus
func NewAS5600(bus drivers.I2C) AS5600Device {
	// Create base device
	baseDev := newBaseDevice(bus)
	// Add AS5600 specific registers
	baseDev.registers[MPOS] = newI2CRegister(MPOS, 0, 0xfff, 2, reg_read|reg_write|reg_program)
	baseDev.registers[MANG] = newI2CRegister(MANG, 0, 0xfff, 2, reg_read|reg_write|reg_program)
	// Add AS5600 specific 'virtual registers'
	conf, ok := baseDev.registers[CONF]
	if ok {
		baseDev.registers[PWMF] = newVirtualRegister(conf, 6, 0b11)
		baseDev.registers[OUTS] = newVirtualRegister(conf, 4, 0b11)
	}
	// Return the device
	return AS5600Device{baseDev}
}

// Configure sets up the AMS AS5600 sensor device with the given configuration.
func (d *AS5600Device) Configure(cfg Config) error {
	// Call the BaseDevice method to do the actual Configure
	d.BaseDevice.Configure(cfg)
	// For AS5600 devices we need to calculate the maxAngle on startup from ZPOS/MPOS/MANG
	// These could have been permanently BURN'ed (by writing BURN register with BURN_ANGLE/BURN_SETTING)
	// or may have already been written in previous runs without a power cycle since.
	mpos, err := d.ReadRegister(MPOS)
	if nil != err {
		return err
	}
	mang, err := d.ReadRegister(MANG)
	if nil != err {
		return err
	}
	// Read ZPOS for side effect of caching only so that next calculateEffectiveMaxAngle() can't fail
	if _, err = d.ReadRegister(ZPOS); nil != err {
		return err
	}
	if mpos != 0 {
		// If MPOS is set, use MPOS regardless of MANG
		err = d.calculateEffectiveMaxAngle(MPOS, mpos)
	} else if mang != 0 {
		// If MANG is set and MPOS == 0, use MANG
		err = d.calculateEffectiveMaxAngle(MANG, mang)
	} else {
		// if neither is set, we have no narrow range
		d.maxAngle = NATIVE_ANGLE_RANGE
	}
	return err
}

// calculateEffectiveMaxAngle calculates d.maxAngle after one of ZPOS/MPOS/MANG have been written
func (d *AS5600Device) calculateEffectiveMaxAngle(register uint8, value uint16) error {

	var zpos, mpos uint16 = 0, 0
	var err error = nil

	switch register {
	case MANG:
		d.maxAngle = value // The easy case
		return nil
	case ZPOS:
		zpos = value
		mpos, err = d.ReadRegister(MPOS)
	case MPOS:
		mpos = value
		zpos, err = d.ReadRegister(ZPOS)
	default:
		panic("calculateEffectiveMaxAngle() can only work from ZPOS, MPOS or MANG")
	}

	if nil != err {
		return err
	}
	// MANG is effectively MPOS-ZPOS
	mang := int(mpos) - int(zpos)
	// correct for mpos < zpos
	if mang < 0 {
		mang += NATIVE_ANGLE_RANGE
	}
	d.maxAngle = uint16(mang)
	return nil
}

// WriteRegister writes the given value for the given register to the AS560x device via I2C
func (d *AS5600Device) WriteRegister(address uint8, value uint16) error {
	// Call the BaseDevice method to do the actual write
	if err := d.BaseDevice.WriteRegister(address, value); err != nil {
		return err
	}
	// When either ZPOS/MANG/MPOS are set we need to recalculate maxAngle
	// We also may need to invalidate some cached values for the other two registers
	recalc := false
	switch address {
	case ZPOS:
		// Setting a new ZPOS invalidates MPOS but not MANG
		d.registers[MPOS].invalidate()
		recalc = true
	case MPOS:
		// Setting a new MPOS invalidates MANG but not ZPOS
		d.registers[MANG].invalidate()
		recalc = true
	case MANG:
		// Setting a new MANG invalidates MPOS but not ZPOS
		d.registers[MPOS].invalidate()
		recalc = true
	}
	if recalc {
		// Datasheet tells us to wait at least 1ms before reading back
		time.Sleep(time.Millisecond * 10) // conservative wait
		return d.calculateEffectiveMaxAngle(address, value)
	}
	return nil
}

// GetMaxPosition returns the 'max position' (MPOS) in different units
func (d *AS5600Device) GetMaxPosition(units AngleUnit) (uint16, float32, error) {
	mpos, err := d.ReadRegister(MPOS)
	if nil != err {
		return 0, 0.0, err
	}
	// Convert to requested units
	i, f := convertFromNativeAngle(mpos, NATIVE_ANGLE_RANGE, units)
	return i, f, nil
}

// SetMaxPosition sets the 'max position' (MPOS) in different units
func (d *AS5600Device) SetMaxPosition(mpos float32, units AngleUnit) error {
	return d.WriteRegister(MPOS, convertToNativeAngle(mpos, units))
}

// GetMaxAngle returns the 'max position' (MANG) in different units
func (d *AS5600Device) GetMaxAngle(units AngleUnit) (uint16, float32, error) {
	mang, err := d.ReadRegister(MANG)
	if nil != err {
		return 0, 0.0, err
	}
	// Convert to requested units
	i, f := convertFromNativeAngle(mang, NATIVE_ANGLE_RANGE, units)
	return i, f, nil
}

// SetMaxAngle sets the 'max angle' (MANG) in different units
func (d *AS5600Device) SetMaxAngle(mang float32, units AngleUnit) error {
	return d.WriteRegister(MANG, convertToNativeAngle(mang, units))
}
