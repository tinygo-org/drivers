// Package veml6070 provides a driver for the VEML6070 digital UV light sensor
// by Vishay.
//
// Datasheet:
// https://www.vishay.com/docs/84277/veml6070.pdf
// Application Notes:
// https://www.vishay.com/docs/84310/designingveml6070.pdf
//
package veml6070 // import "tinygo.org/x/drivers/veml6070"

import (
	"time"

	"tinygo.org/x/drivers"
)

// Device wraps an I2C connection to a VEML6070 device.
type Device struct {
	bus         drivers.I2C
	AddressLow  uint16
	AddressHigh uint16
	RSET        uint32
	IT          uint8
}

// New creates a new VEML6070 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not initialize the device.
// You must call Configure() first in order to use the device itself.
func New(bus drivers.I2C) Device {
	return Device{
		bus:         bus,
		AddressLow:  ADDR_L,
		AddressHigh: ADDR_H,
		RSET:        RSET_240K,
		// Note: default to maximum to get as much precision as possible since
		// raw data values larger than 16 bit can hardly occur with RSET below
		// 300 kOhm in real world applications. Power saving due to shorter
		// sampling time might be a reason to reduce this.
		IT: IT_4,
	}
}

// Configure sets up the device for communication
func (d *Device) Configure() bool {
	// save power by shutdown as early as possible, also serves as presence test
	if err := d.disable(); err != nil {
		return false
	}

	return true
}

// ReadUVALightIntensity returns the UVA light intensity (irradiance)
// in milli Watt per square meter (mW/(m*m))
func (d *Device) ReadUVALightIntensity() (uint32, error) {
	var err2 error

	if err := d.enable(); err != nil {
		return 0, err
	}

	// wait two times the refresh time to allow completion of a previous cycle
	// with old settings (worst case)
	time.Sleep(time.Duration(d.getRefreshTime()) * 2 * time.Millisecond)

	msb, err2 := d.readData(d.AddressHigh)
	if err2 != nil {
		return 0, err2
	}

	lsb, err2 := d.readData(d.AddressLow)
	if err2 != nil {
		return 0, err2
	}

	if err := d.disable(); err != nil {
		return 0, err
	}

	rawData := (uint32(msb) << 8) | uint32(lsb)

	// normalize raw data (step count sampled in d.getRefreshTime()) into the
	// linearly scaled normalized data (step count sampled in 100ms) for which
	// we know the UVA sensitivity
	normalizedData := float32(rawData) * NORMALIZED_REFRESHTIME / d.getRefreshTime()

	// now we can calculate the absolute UVA power detected combining normalized
	// data with known UVA sensitivity for this data, from datasheet
	intensity := normalizedData * NORMALIZED_UVA_SENSITIVITY // mW/(m*m)

	return uint32(intensity + 0.5), nil
}

// GetEstimatedRiskLevel returns estimated risk level from comparing UVA light
// intensity values in mW/(m*m) with thresholds calculated from application notes
func (d *Device) GetEstimatedRiskLevel(intensity uint32) uint8 {
	if intensity <= 24888 {
		return UVI_RISK_LOW
	} else if intensity <= 49800 {
		return UVI_RISK_MODERATE
	} else if intensity <= 66400 {
		return UVI_RISK_HIGH
	} else if intensity <= 91288 {
		return UVI_RISK_VERY_HIGH
	} else {
		return UVI_RISK_EXTREME
	}
}

func (d *Device) disable() error {
	return d.bus.Tx(uint16(d.AddressLow), []byte{CONFIG_DISABLE}, nil)
}

func (d *Device) enable() error {
	return d.bus.Tx(uint16(d.AddressLow), []byte{CONFIG_ENABLE | d.IT}, nil)
}

func (d *Device) readData(address uint16) (byte, error) {
	data := []byte{0}
	err := d.bus.Tx(address, []byte{}, data)
	return data[0], err
}

// getRefreshTime returns the refresh time (aka sample time) in milliseconds
func (d *Device) getRefreshTime() float32 {
	var it float32
	switch d.IT {
	case IT_HALF:
		it = 0.5
	case IT_1:
		it = 1
	case IT_2:
		it = 2
	case IT_4:
		it = 4
	}
	return float32(d.RSET) * RSET_TO_REFRESHTIME_SCALE * it
}
