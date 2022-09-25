// Package vl6180x provides a driver for the VL6180X time-of-flight distance sensor
//
// Datasheet:
// https://www.st.com/resource/en/datasheet/vl6180x.pdf
// This driver was based on the library https://github.com/adafruit/Adafruit_VL6180X
// and document 'AN4545 VL6180X basic ranging application note':
// https://www.st.com/resource/en/application_note/an4545-vl6180x-basic-ranging-application-note-stmicroelectronics.pdf
package vl6180x // import "tinygo.org/x/drivers/vl6180x"

import (
	"time"

	"tinygo.org/x/drivers"
)

type VL6180XError uint8

// Device wraps an I2C connection to a VL6180X device.
type Device struct {
	bus     drivers.I2C
	Address uint16
	timeout uint32
}

// New creates a new VL6180X connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
		timeout: 500,
	}
}

// Connected returns whether a VL6180X has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {
	return d.readReg(WHO_AM_I) == CHIP_ID
}

// Configure sets up the device for communication
func (d *Device) Configure(use2v8Mode bool) bool {
	if !d.Connected() {
		return false
	}

	if (d.readReg(SYSTEM_FRESH_OUT_OF_RESET) & 0x01) == 0x01 {

		// mandatory settings from page 24 of AN4545
		d.writeReg(0x0207, 0x01)
		d.writeReg(0x0208, 0x01)
		d.writeReg(0x0096, 0x00)
		d.writeReg(0x0097, 0xfd)
		d.writeReg(0x00e3, 0x00)
		d.writeReg(0x00e4, 0x04)
		d.writeReg(0x00e5, 0x02)
		d.writeReg(0x00e6, 0x01)
		d.writeReg(0x00e7, 0x03)
		d.writeReg(0x00f5, 0x02)
		d.writeReg(0x00d9, 0x05)
		d.writeReg(0x00db, 0xce)
		d.writeReg(0x00dc, 0x03)
		d.writeReg(0x00dd, 0xf8)
		d.writeReg(0x009f, 0x00)
		d.writeReg(0x00a3, 0x3c)
		d.writeReg(0x00b7, 0x00)
		d.writeReg(0x00bb, 0x3c)
		d.writeReg(0x00b2, 0x09)
		d.writeReg(0x00ca, 0x09)
		d.writeReg(0x0198, 0x01)
		d.writeReg(0x01b0, 0x17)
		d.writeReg(0x01ad, 0x00)
		d.writeReg(0x00ff, 0x05)
		d.writeReg(0x0100, 0x05)
		d.writeReg(0x0199, 0x05)
		d.writeReg(0x01a6, 0x1b)
		d.writeReg(0x01ac, 0x3e)
		d.writeReg(0x01a7, 0x1f)
		d.writeReg(0x0030, 0x00)

		// recommended settings
		d.writeReg(0x0011, 0x10) // enables polling when measurement completes
		d.writeReg(0x010a, 0x30) // sets averaging sample period
		d.writeReg(0x003f, 0x46) // sets light and dark gain
		d.writeReg(0x0031, 0xFF) // sets the # of range measurements for auto calibration
		d.writeReg(0x0041, 0x63) // sets ALS integration time to 100ms
		d.writeReg(0x002e, 0x01) // performs a single temperature calibration

		// optional settings
		d.writeReg(RANGING_INTERMEASUREMENT_PERIOD, 0x09) // sets ranging inter-measurement period to 100ms
		d.writeReg(ALS_INTERMEASUREMENT_PERIOD, 0x31)     // sets default ALS inter-measurement period to 500ms
		d.writeReg(SYSTEM_INTERRUPT_CONFIG, 0x24)         // configures interrupt

		d.writeReg(SYSTEM_FRESH_OUT_OF_RESET, 0x00)
		time.Sleep(100 * time.Microsecond)
	}

	return true
}

// Read returns the proximity of the sensor in mm
func (d *Device) Read() uint16 {
	start := time.Now()

	for d.dataReady() {
		elapsed := time.Since(start)
		if d.timeout > 0 && uint32(elapsed.Seconds()*1000) > d.timeout {
			return 0
		}
	}

	d.writeReg(SYSRANGE_START, 0x01)
	for (d.readReg(RESULT_INTERRUPT_STATUS_GPIO) & 0x04) == 0 {
	}

	return uint16(d.readRangeResult())
}

// dataReady returns true when the data is ready to be read
func (d *Device) dataReady() bool {
	return (d.readReg(RESULT_RANGE_STATUS) & 0x01) == 0
}

// startRange starts the readings
func (d *Device) startRange() {
	for d.dataReady() {
	}
	d.writeReg(SYSRANGE_START, 0x01)
}

// IsRangeComplete return true when the reading is complete
func (d *Device) IsRangeComplete() bool {
	if (d.readReg(RESULT_INTERRUPT_STATUS_GPIO) & 0x04) != 0 {
		return true
	}
	return false
}

// readRangeResults returns the sensor value from the register
func (d *Device) readRangeResult() uint8 {
	value := d.readReg(RESULT_RANGE_VAL)

	d.writeReg(SYSTEM_INTERRUPT_CLEAR, 0x07)
	return value
}

// StartRangeContinuous starts the continuous reading mode
func (d *Device) StartRangeContinuous(periodInMs uint16) {
	var periodReg uint8
	if periodInMs > 10 {
		if periodInMs < 2550 {
			periodReg = uint8(periodInMs/10) - 1
		} else {
			periodReg = 254
		}
	}
	d.writeReg(RANGING_INTERMEASUREMENT_PERIOD, periodReg)
	d.writeReg(SYSRANGE_START, 0x03)
}

// StopRangeContinuous stops the continuous reading mode
func (d *Device) StopRangeContinuous() {
	d.writeReg(SYSRANGE_START, 0x01)
}

// ReadStatus returns the current status of the sensor
func (d *Device) ReadStatus() uint8 {
	return d.readReg(RESULT_RANGE_STATUS) >> 4
}

// ReadLux returns the lux of the sensor
func (d *Device) ReadLux(gain uint8) (lux uint32) {
	reg := d.readReg(SYSTEM_INTERRUPT_CONFIG)
	reg &= ^uint8(0x38)
	reg |= 0x4 << 3
	d.writeReg(SYSTEM_INTERRUPT_CONFIG, reg)

	d.writeReg(SYSALS_INTEGRATION_PERIOD_HI, 0)
	d.writeReg(SYSALS_INTEGRATION_PERIOD_HI, 100)

	if gain > ALS_GAIN_40 {
		gain = ALS_GAIN_40
	}
	d.writeReg(SYSALS_ANALOGUE_GAIN, 0x40|gain)

	d.writeReg(SYSALS_START, 0x1)
	for 4 != ((d.readReg(RESULT_INTERRUPT_STATUS_GPIO) >> 3) & 0x7) {
	}

	lux = uint32(d.readReg16Bit(RESULT_ALS_VAL)) * 320
	d.writeReg(SYSTEM_INTERRUPT_CLEAR, 0x07)

	switch gain {
	case ALS_GAIN_1:
		break
	case ALS_GAIN_1_25:
		lux = (lux * 100) / 125
		break
	case ALS_GAIN_1_67:
		lux = (lux * 100) / 167
		break
	case ALS_GAIN_2_5:
		lux = (lux * 10) / 25
		break
	case ALS_GAIN_5:
		lux /= 5
		break
	case ALS_GAIN_10:
		lux /= 10
		break
	case ALS_GAIN_20:
		lux /= 20
		break
	case ALS_GAIN_40:
		lux /= 40
		break
	}

	return lux
}

// SetOffset sets the offset
func (d *Device) SetOffset(offset uint8) {
	d.writeReg(SYSRANGE_PART_TO_PART_RANGE_OFFSET, offset)
}

// SetAddress sets the I2C address which this device listens to.
func (d *Device) SetAddress(address uint8) {
	d.writeReg(I2C_SLAVE_DEVICE_ADDRESS, address)
	d.Address = uint16(address)
}

// GetAddress returns the I2C address which this device listens to.
func (d *Device) GetAddress() uint8 {
	return uint8(d.Address)
}

// writeReg sends a single byte to the specified register address
func (d *Device) writeReg(reg uint16, value uint8) {
	msb := byte((reg >> 8) & 0xFF)
	lsb := byte(reg & 0xFF)
	d.bus.Tx(d.Address, []byte{msb, lsb, value}, nil)
}

// readReg reads a single byte from the specified address
func (d *Device) readReg(reg uint16) uint8 {
	data := []byte{0}
	msb := byte((reg >> 8) & 0xFF)
	lsb := byte(reg & 0xFF)
	d.bus.Tx(d.Address, []byte{msb, lsb}, data)
	return data[0]
}

// readReg16Bit reads two bytes from the specified address
// and returns it as a uint16
func (d *Device) readReg16Bit(reg uint16) uint16 {
	data := []byte{0, 0}
	msb := byte((reg >> 8) & 0xFF)
	lsb := byte(reg & 0xFF)
	d.bus.Tx(d.Address, []byte{msb, lsb}, data)
	return readUint(data[0], data[1])
}

// readUint converts two bytes to uint16
func readUint(msb byte, lsb byte) uint16 {
	return (uint16(msb) << 8) | uint16(lsb)
}
