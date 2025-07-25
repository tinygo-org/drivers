// Package ens160 provides a driver for the ScioSense ENS160 digital gas sensor.
//
// Datasheet: https://www.sciosense.com/wp-content/uploads/2023/12/ENS160-Datasheet.pdf
package ens160

import (
	"encoding/binary"
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

const (
	defaultTimeout = 30 * time.Millisecond
	shortTimeout   = 1 * time.Millisecond
)

// Conversion constants for environment data compensation.
const (
	kelvinOffsetMilli = 273150          // 273.15 K in milli-units
	tempRawFactor     = 64              // As per datasheet for TEMP_IN
	humRawFactor      = 512             // As per datasheet for RH_IN
	milliFactor       = 1000            // For converting from milli-units
	roundingTerm      = milliFactor / 2 // For rounding before integer division
)

// validityStrings provides human-readable descriptions for validity flags.
var validityStrings = [...]string{
	ValidityNormalOperation:     "normal operation",
	ValidityWarmUpPhase:         "warm-up phase, wait ~3 minutes for valid data",
	ValidityInitialStartUpPhase: "initial start-up phase, wait ~1 hour for valid data",
	ValidityInvalidOutput:       "invalid output",
}

// Device wraps an I2C connection to an ENS160 device.
type Device struct {
	bus  drivers.I2C // I²C implementation
	addr uint16      // 7‑bit bus address, promoted to uint16 per drivers.I2C

	// shadow registers / last measurements
	lastTvocPPB  uint16
	lastEco2PPM  uint16
	lastAqiUBA   uint8
	lastValidity uint8 // Store the latest validity status

	// pre‑allocated buffers
	wbuf [5]byte // longest write: reg + 4 bytes (TEMP+RH)
	rbuf [5]byte // longest read: DATA burst (5 bytes)
}

// New returns a new ENS160 driver.
func New(bus drivers.I2C, addr uint16) *Device {
	if addr == 0 {
		addr = DefaultAddress
	}
	return &Device{
		bus:          bus,
		addr:         addr,
		lastValidity: ValidityInvalidOutput,
	}
}

// Connected returns whether a ENS160 has been found.
func (d *Device) Connected() bool {
	d.wbuf[0] = regPartID
	err := d.bus.Tx(d.addr, d.wbuf[:1], d.rbuf[:2])
	return err == nil && d.rbuf[0] == LowPartID && d.rbuf[1] == HighPartID
}

// Configure sets up the device for reading.
func (d *Device) Configure() error {
	// 1. Soft-reset. The device will automatically enter IDLE mode.
	if err := d.write1(regOpMode, ModeReset); err != nil {
		return err
	}
	time.Sleep(defaultTimeout)

	// 2. Clear GPR registers, then go to STANDARD mode.
	if err := d.write1(regCommand, cmdClrGPR); err != nil {
		return err
	}
	time.Sleep(defaultTimeout)

	if err := d.write1(regOpMode, ModeStandard); err != nil {
		return err
	}
	time.Sleep(defaultTimeout)

	return nil
}

// calculateTempRaw converts temperature from milli-degrees Celsius to the sensor's raw format.
func calculateTempRaw(tempMilliC int32) uint16 {
	// Clip temperature
	const (
		minC = -40 * 1000
		maxC = 85 * 1000
	)
	if tempMilliC < minC {
		tempMilliC = minC
	} else if tempMilliC > maxC {
		tempMilliC = maxC
	}

	// Integer fixed-point conversion to format required by the sensor.
	// Formula from datasheet: T_IN = (T_ambient_C + 273.15) * 64
	return uint16((((tempMilliC + kelvinOffsetMilli) * tempRawFactor) + roundingTerm) / milliFactor)
}

// calculateHumRaw converts relative humidity from milli-percent to the sensor's raw format.
func calculateHumRaw(rhMilliPct int32) uint16 {
	// Clip humidity
	if rhMilliPct < 0 {
		rhMilliPct = 0
	} else if rhMilliPct > 100*1000 {
		rhMilliPct = 100 * 1000
	}

	// Integer fixed-point conversion to format required by the sensor.
	// Formula from datasheet: RH_IN = (RH_ambient_% * 512)
	return uint16(((rhMilliPct * humRawFactor) + roundingTerm) / milliFactor)
}

// SetEnvDataMilli sets the ambient temperature and humidity for compensation.
//
// tempMilliC is the temperature in milli-degrees Celsius.
// rhMilliPct is the relative humidity in milli-percent.
func (d *Device) SetEnvDataMilli(tempMilliC, rhMilliPct int32) error {
	tempRaw := calculateTempRaw(tempMilliC)
	humRaw := calculateHumRaw(rhMilliPct)

	d.wbuf[0] = regTempIn // start address (auto‑increment)
	binary.LittleEndian.PutUint16(d.wbuf[1:3], tempRaw)
	binary.LittleEndian.PutUint16(d.wbuf[3:5], humRaw)

	return d.bus.Tx(d.addr, d.wbuf[:5], nil)
}

// Update refreshes the concentration measurements.
func (d *Device) Update(which drivers.Measurement) error {
	if which&drivers.Concentration == 0 {
		return nil // nothing requested
	}

	const maxTries = 1000
	var (
		status   uint8
		validity uint8
	)
	var gotData bool

	// Poll DEVICE_STATUS until NEWDAT or timeout
	for range maxTries {
		var err error
		status, err = d.read1(regStatus)
		if err != nil {
			return err
		}
		if status&statusSTATER != 0 {
			return errors.New("ENS160: error (STATER set)")
		}
		validity = (status & statusValidityMask) >> statusValidityShift

		if status&statusNEWDAT != 0 {
			gotData = true
			break // Always break when data available
		}
		time.Sleep(shortTimeout)
	}
	if !gotData {
		return errors.New("ENS160: timeout waiting for NEWDAT")
	}

	// Burst-read data regardless of validity state
	d.wbuf[0] = regAQI
	if err := d.bus.Tx(d.addr, d.wbuf[:1], d.rbuf[:5]); err != nil {
		return errors.New("ENS160: burst read failed")
	}

	d.lastAqiUBA = d.rbuf[0]
	d.lastTvocPPB = binary.LittleEndian.Uint16(d.rbuf[1:3])
	d.lastEco2PPM = binary.LittleEndian.Uint16(d.rbuf[3:5])
	d.lastValidity = validity // Store the validity status

	return nil
}

// TVOC returns the last total‑VOC concentration in parts‑per‑billion.
func (d *Device) TVOC() uint16 { return d.lastTvocPPB }

// ECO2 returns the last equivalent CO₂ concentration in parts‑per‑million.
func (d *Device) ECO2() uint16 { return d.lastEco2PPM }

// AQI returns the last Air‑Quality Index according to UBA (1–5).
func (d *Device) AQI() uint8 { return d.lastAqiUBA }

// Validity returns the current operating state of the sensor.
func (d *Device) Validity() uint8 {
	return d.lastValidity
}

// ValidityString returns a human-readable string describing the current validity status.
func (d *Device) ValidityString() string {
	if int(d.lastValidity) < len(validityStrings) {
		return validityStrings[d.lastValidity]
	}
	return "unknown"
}

// write1 writes a single byte to a register.
func (d *Device) write1(reg, val uint8) error {
	d.wbuf[0] = reg
	d.wbuf[1] = val
	return d.bus.Tx(d.addr, d.wbuf[:2], nil)
}

// read1 reads a single byte from a register.
func (d *Device) read1(reg uint8) (uint8, error) {
	d.wbuf[0] = reg
	if err := d.bus.Tx(d.addr, d.wbuf[:1], d.rbuf[:1]); err != nil {
		return 0, err
	}
	return d.rbuf[0], nil
}
