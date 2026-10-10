// Package bh1745 implements a driver for the bh1745 color sensor
// It detects 16-bit RGBC (Red, Green, Blue, Clear) color over
// I2C interface.
//
// Datasheet: https://www.mouser.co.uk/datasheet/2/348/bh1745nuc-e-519994.pdf
//
// Pimoroni bh1745-python driver
// https://github.com/pimoroni/bh1745-python/blob/main/bh1745/__init__.py

package bh1745 // import "tinygo.org/x/drivers/bh1745"

import (
	"math"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// RGBCData holds one set of raw channel readings.
type RGBCData struct {
	R, G, B, C uint16
}

// Device is the BH1745 RGBC colour sensor driver.
type Device struct {
	bus     drivers.I2C
	Address uint16

	r, g, b, c uint16 // cached raw RGBC readings from the last Update call
}

// New creates a new BH1745 driver. The I2C bus must already be configured.
// Call Configure() before calling Read().
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
	}
}

// Configure resets the sensor and sets it up with default settings:
// 160 ms measurement time, 1× ADC gain, RGBC measurement enabled.
//
// The undocumented write of 0x02 to MODE_CONTROL3 (register 0x44) is included
// here; without it the sensor reports incorrect RGBC values. This matches the
// workaround applied in Pimoroni's Enviro Indoor firmware.
func (d *Device) Configure() error {
	// Software reset.
	if err := legacy.WriteRegister(d.bus, uint8(d.Address), REG_SYSTEM_CONTROL, []byte{SW_RESET}); err != nil {
		return err
	}

	// Wait for sw_reset to self-clear (typically < 1 ms; allow up to 200 ms).
	for i := 0; i < 200; i++ {
		var buf [1]byte
		if err := legacy.ReadRegister(d.bus, uint8(d.Address), REG_SYSTEM_CONTROL, buf[:]); err != nil {
			return err
		}
		if buf[0]&SW_RESET == 0 {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	// Clear int_reset flag.
	if err := legacy.WriteRegister(d.bus, uint8(d.Address), REG_SYSTEM_CONTROL, []byte{0x00}); err != nil {
		return err
	}

	// 160 ms measurement time.
	if err := legacy.WriteRegister(d.bus, uint8(d.Address), REG_MODE_CONTROL1, []byte{byte(MeasTime160ms)}); err != nil {
		return err
	}

	// Enable RGBC with 1× gain.
	if err := legacy.WriteRegister(d.bus, uint8(d.Address), REG_MODE_CONTROL2, []byte{RGBC_EN | byte(Gain1X)}); err != nil {
		return err
	}

	// Undocumented workaround: writing 0x02 to MODE_CONTROL3 is required for
	// the sensor to report correct colour data. Must be applied after reset.
	if err := legacy.WriteRegister(d.bus, uint8(d.Address), REG_MODE_CONTROL3, []byte{0x02}); err != nil {
		return err
	}

	// Wait for one full 160 ms measurement cycle before the first read.
	time.Sleep(160 * time.Millisecond)
	return nil
}

// Connected returns true when the sensor is present and the part ID and
// manufacturer ID both match expected values.
func (d *Device) Connected() bool {
	var buf [1]byte
	legacy.ReadRegister(d.bus, uint8(d.Address), REG_SYSTEM_CONTROL, buf[:])
	if buf[0]&PART_ID_MASK != PART_ID {
		return false
	}
	legacy.ReadRegister(d.bus, uint8(d.Address), REG_MANUFACTURER, buf[:])
	return buf[0] == MANUFACTURER_ID
}

// SetMeasurementTime changes the RGBC integration time. Wait at least one full
// cycle at the new time before calling Read() again.
func (d *Device) SetMeasurementTime(t MeasurementTime) {
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_MODE_CONTROL1, []byte{byte(t) & 0x07})
}

// SetADCGain changes the ADC gain multiplier.
func (d *Device) SetADCGain(g ADCGain) {
	var buf [1]byte
	legacy.ReadRegister(d.bus, uint8(d.Address), REG_MODE_CONTROL2, buf[:])
	buf[0] = (buf[0] &^ 0x03) | byte(g)
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_MODE_CONTROL2, buf[:])
}

// Update reads a fresh set of RGBC values from the sensor and caches them.
// Pass drivers.Luminosity (or drivers.AllMeasurements) to trigger a read.
// The cached values are then available via R(), G(), B(), C(), Luminosity(),
// and ColorTemperature().
func (d *Device) Update(which drivers.Measurement) error {
	if which&drivers.Luminosity == 0 {
		return nil
	}
	var buf [8]byte
	if err := legacy.ReadRegister(d.bus, uint8(d.Address), REG_RED_DATA_LSB, buf[:]); err != nil {
		return err
	}
	d.r = uint16(buf[1])<<8 | uint16(buf[0])
	d.g = uint16(buf[3])<<8 | uint16(buf[2])
	d.b = uint16(buf[5])<<8 | uint16(buf[4])
	d.c = uint16(buf[7])<<8 | uint16(buf[6])
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

// Read returns the last cached raw RGBC channel values without performing I2C.
// Call Update first to refresh the values.
//
// Deprecated: call Update then use R(), G(), B(), C() instead.
func (d *Device) Read() RGBCData {
	return RGBCData{R: d.r, G: d.g, B: d.b, C: d.c}
}

// CalcLux calculates the illuminance in lux from raw RGBC values.
//
// Formula ported from Pimoroni's Enviro Indoor firmware (indoor.py), tuned for
// the BH1745 configured at 160 ms integration time and 1× gain.
func CalcLux(r, g, b, c uint16) uint32 {
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
	rf := float64(r)
	gf := float64(g)
	bf := float64(b)
	cf := float64(c)
	sum := rf + gf + bf
	if sum < 1 {
		return 0
	}

	rRatio := rf / sum
	bRatio := bf / sum

	var ct float64
	if cf/gf < 0.160 {
		bEff := bRatio * 3.13
		if bEff > 1 {
			bEff = 1
		}
		ct = (1-bEff)*12746*math.Exp(-2.911*rRatio) + bEff*1637*math.Exp(4.865*bRatio)
	} else {
		bEff := bRatio * 10.67
		if bEff > 1 {
			bEff = 1
		}
		ct = (1-bEff)*16234*math.Exp(-2.781*rRatio) + bEff*1882*math.Exp(4.448*bRatio)
	}
	if ct > 10000 {
		ct = 10000
	}
	return uint32(ct)
}
