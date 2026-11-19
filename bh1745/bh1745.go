// Package bh1745 provides a driver for the BH1745NUC RGBC colour sensor by ROHM.
//
// Datasheet: https://fscdn.rohm.com/en/products/databook/datasheet/ic/sensor/light/bh1745nuc-e.pdf
package bh1745 // import "tinygo.org/x/drivers/bh1745"

import (
	"math"
	"time"

	"tinygo.org/x/drivers"
)

// Device is the BH1745 RGBC colour sensor driver.
type Device struct {
	bus     drivers.I2C
	Address uint16
	wbuf    [2]byte
	rbuf    [8]byte

	r, g, b, c uint16 // cached raw RGBC readings from the last Update call
}

// New creates a new BH1745 driver. The I2C bus must already be configured.
// Call Configure() before calling Update().
func New(bus drivers.I2C) Device {
	return NewWithAddress(bus, Address)
}

// NewWithAddress creates a new BH1745 driver at the given I2C address.
// Use AddressAlt when the ADDR pin is high.
func NewWithAddress(bus drivers.I2C, address uint16) Device {
	return Device{
		bus:     bus,
		Address: address,
	}
}

func (d *Device) readRegister(reg byte, buf []byte) error {
	d.wbuf[0] = reg
	return d.bus.Tx(d.Address, d.wbuf[:1], buf)
}

func (d *Device) writeRegister(reg, value byte) error {
	d.wbuf[0] = reg
	d.wbuf[1] = value
	return d.bus.Tx(d.Address, d.wbuf[:], nil)
}

// Configure resets the sensor and sets it up with default settings:
// 160 ms measurement time, 1× ADC gain, RGBC measurement enabled.
//
// The undocumented write of 0x02 to MODE_CONTROL3 (register 0x44) is included
// here; without it the sensor reports incorrect RGBC values. This matches the
// workaround applied in Pimoroni's Enviro Indoor firmware.
func (d *Device) Configure() error {
	// Software reset.
	if err := d.writeRegister(REG_SYSTEM_CONTROL, SW_RESET); err != nil {
		return err
	}

	// Wait for sw_reset to self-clear (typically < 1 ms; allow up to 200 ms).
	for i := 0; i < 200; i++ {
		if err := d.readRegister(REG_SYSTEM_CONTROL, d.rbuf[:1]); err != nil {
			return err
		}
		if d.rbuf[0]&SW_RESET == 0 {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	// Clear int_reset flag.
	if err := d.writeRegister(REG_SYSTEM_CONTROL, 0x00); err != nil {
		return err
	}

	// 160 ms measurement time.
	if err := d.writeRegister(REG_MODE_CONTROL1, byte(MeasTime160ms)); err != nil {
		return err
	}

	// Enable RGBC with 1× gain.
	if err := d.writeRegister(REG_MODE_CONTROL2, RGBC_EN|byte(Gain1X)); err != nil {
		return err
	}

	// Undocumented workaround: writing 0x02 to MODE_CONTROL3 is required for
	// the sensor to report correct colour data. Must be applied after reset.
	if err := d.writeRegister(REG_MODE_CONTROL3, 0x02); err != nil {
		return err
	}

	// Wait for one full 160 ms measurement cycle before the first read.
	time.Sleep(160 * time.Millisecond)
	return nil
}

// Connected returns true when the sensor is present and the part ID and
// manufacturer ID both match expected values.
func (d *Device) Connected() bool {
	if err := d.readRegister(REG_SYSTEM_CONTROL, d.rbuf[:1]); err != nil {
		return false
	}
	if d.rbuf[0]&PART_ID_MASK != PART_ID {
		return false
	}
	if err := d.readRegister(REG_MANUFACTURER, d.rbuf[:1]); err != nil {
		return false
	}
	return d.rbuf[0] == MANUFACTURER_ID
}

// SetMeasurementTime changes the RGBC integration time. Wait at least one full
// cycle at the new time before calling Update() again. It returns an I2C error
// if the register write fails.
func (d *Device) SetMeasurementTime(t MeasurementTime) error {
	return d.writeRegister(REG_MODE_CONTROL1, byte(t)&0x07)
}

// SetADCGain changes the ADC gain multiplier. It returns an I2C error if the
// read-modify-write fails.
func (d *Device) SetADCGain(g ADCGain) error {
	if err := d.readRegister(REG_MODE_CONTROL2, d.rbuf[:1]); err != nil {
		return err
	}
	d.rbuf[0] = (d.rbuf[0] &^ 0x03) | byte(g)
	return d.writeRegister(REG_MODE_CONTROL2, d.rbuf[0])
}

// Update reads a fresh set of RGBC values from the sensor and caches them.
// Pass drivers.Luminosity (or drivers.AllMeasurements) to trigger a read.
// The cached values are then available via R(), G(), B(), C(), Luminosity(),
// and ColorTemperature().
func (d *Device) Update(which drivers.Measurement) error {
	if which&drivers.Luminosity == 0 {
		return nil
	}
	if err := d.readRegister(REG_RED_DATA_LSB, d.rbuf[:]); err != nil {
		return err
	}
	d.r = uint16(d.rbuf[1])<<8 | uint16(d.rbuf[0])
	d.g = uint16(d.rbuf[3])<<8 | uint16(d.rbuf[2])
	d.b = uint16(d.rbuf[5])<<8 | uint16(d.rbuf[4])
	d.c = uint16(d.rbuf[7])<<8 | uint16(d.rbuf[6])
	return nil
}

// R returns the cached raw red channel value from the last Update call.
func (d *Device) R() uint16 { return d.r }

// G returns the cached raw green channel value from the last Update call.
func (d *Device) G() uint16 { return d.g }

// B returns the cached raw blue channel value from the last Update call.
func (d *Device) B() uint16 { return d.b }

// C returns the cached raw clear channel value from the last Update call.
func (d *Device) C() uint16 { return d.c }

// Luminosity returns the illuminance in lux calculated from the cached RGBC
// values. Call Update first.
func (d *Device) Luminosity() uint32 {
	return CalcLux(d.r, d.g, d.b, d.c)
}

// ColorTemperature returns the correlated colour temperature in Kelvin
// calculated from the cached RGBC values. Call Update first.
func (d *Device) ColorTemperature() uint32 {
	return CalcColorTemperature(d.r, d.g, d.b, d.c)
}

// CalcLux calculates the illuminance in lux from raw RGBC values.
//
// Formula ported from Pimoroni's Enviro Indoor firmware (indoor.py), tuned for
// the BH1745 configured at 160 ms integration time and 1× gain.
func CalcLux(r, g, b, c uint16) uint32 {
	// b is accepted for symmetry with CalcColorTemperature; the Pimoroni lux
	// formula uses only r, g, and c.
	if g < 1 {
		return 0
	}
	rf := float32(r)
	gf := float32(g)
	cf := float32(c)
	var tmp float32
	if cf/gf < 0.160 {
		tmp = 0.202*rf + 0.766*gf
	} else {
		tmp = 0.159*rf + 0.646*gf
	}
	if tmp < 0 {
		return 0
	}
	return uint32(tmp)
}

// CalcColorTemperature calculates the correlated colour temperature (CCT) in Kelvin
// from raw RGBC values.
//
// Formula ported from Pimoroni's Enviro Indoor firmware (indoor.py).
// Returns 0 when the light level is too low for a reliable estimate.
func CalcColorTemperature(r, g, b, c uint16) uint32 {
	if g < 1 {
		return 0
	}
	rf := float32(r)
	gf := float32(g)
	bf := float32(b)
	cf := float32(c)
	sum := rf + gf + bf
	if sum < 1 {
		return 0
	}

	rRatio := rf / sum
	bRatio := bf / sum

	var ct float32
	if cf/gf < 0.160 {
		bEff := bRatio * 3.13
		if bEff > 1 {
			bEff = 1
		}
		ct = (1-bEff)*12746*float32(math.Exp(float64(-2.911*rRatio))) +
			bEff*1637*float32(math.Exp(float64(4.865*bRatio)))
	} else {
		bEff := bRatio * 10.67
		if bEff > 1 {
			bEff = 1
		}
		ct = (1-bEff)*16234*float32(math.Exp(float64(-2.781*rRatio))) +
			bEff*1882*float32(math.Exp(float64(4.448*bRatio)))
	}
	if ct > 10000 {
		ct = 10000
	}
	return uint32(ct)
}
