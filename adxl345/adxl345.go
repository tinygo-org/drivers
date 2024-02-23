// Package adxl345 provides a driver for the ADXL345 digital accelerometer.
//
// Datasheet EN: http://www.analog.com/media/en/technical-documentation/data-sheets/ADXL345.pdf
//
// Datasheet JP: http://www.analog.com/media/jp/technical-documentation/data-sheets/ADXL345_jp.pdf
package adxl345 // import "tinygo.org/x/drivers/adxl345"

import (
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

type Range uint8
type Rate uint8

// Internal structure for the power configuration
type powerCtl struct {
	link      uint8
	autoSleep uint8
	measure   uint8
	sleep     uint8
	wakeUp    uint8
}

// Internal structure for the sensor's data format configuration
type dataFormat struct {
	selfTest    uint8
	spi         uint8
	intInvert   uint8
	fullRes     uint8
	justify     uint8
	sensorRange Range
}

// Internal structure for the sampling rate configuration
type bwRate struct {
	lowPower uint8
	rate     Rate
}

// Device wraps an I2C connection to a ADXL345 device.
type Device struct {
	bus        drivers.I2C
	Address    uint16
	powerCtl   powerCtl
	dataFormat dataFormat
	bwRate     bwRate
}

// New creates a new ADXL345 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not init the device.
// To do that you must call the Configure() method on the Device before using it.
func New(bus drivers.I2C) Device {
	return Device{
		bus: bus,
		powerCtl: powerCtl{
			measure: 1,
		},
		dataFormat: dataFormat{
			sensorRange: RANGE_2G,
		},
		bwRate: bwRate{
			lowPower: 1,
			rate:     RATE_100HZ,
		},
		Address: AddressLow,
	}
}

// Configure sets up the device for communication
func (d *Device) Configure() {
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_BW_RATE, []byte{d.bwRate.toByte()})
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_POWER_CTL, []byte{d.powerCtl.toByte()})
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_DATA_FORMAT, []byte{d.dataFormat.toByte()})
}

// Halt stops the sensor, values will not updated
func (d *Device) Halt() {
	d.powerCtl.measure = 0
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_POWER_CTL, []byte{d.powerCtl.toByte()})
}

// Restart makes reading the sensor working again after a halt
func (d *Device) Restart() {
	d.powerCtl.measure = 1
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_POWER_CTL, []byte{d.powerCtl.toByte()})
}

// ReadAcceleration reads the current acceleration from the device and returns
// it in µg (micro-gravity). When one of the axes is pointing straight to Earth
// and the sensor is not moving the returned value will be around 1000000 or
// -1000000.
func (d *Device) ReadAcceleration() (x int32, y int32, z int32, err error) {
	rx, ry, rz := d.ReadRawAcceleration()

	x = int32(d.dataFormat.convertToIS(rx))
	y = int32(d.dataFormat.convertToIS(ry))
	z = int32(d.dataFormat.convertToIS(rz))

	return
}

// ReadRawAcceleration reads the sensor values and returns the raw x, y and z axis
// from the adxl345.
func (d *Device) ReadRawAcceleration() (x int16, y int16, z int16) {
	data := []byte{0, 0, 0, 0, 0, 0}
	legacy.ReadRegister(d.bus, uint8(d.Address), REG_DATAX0, data)

	x = readIntLE(data[0], data[1])
	y = readIntLE(data[2], data[3])
	z = readIntLE(data[4], data[5])

	return
}

// UseLowPower sets the ADXL345 to use the low power mode.
func (d *Device) UseLowPower(power bool) {
	if power {
		d.bwRate.lowPower = 1
	} else {
		d.bwRate.lowPower = 0
	}
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_BW_RATE, []byte{d.bwRate.toByte()})
}

// SetRate change the current rate of the sensor
func (d *Device) SetRate(rate Rate) bool {
	d.bwRate.rate = rate & 0x0F
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_BW_RATE, []byte{d.bwRate.toByte()})
	return true
}

// SetRange change the current range of the sensor
func (d *Device) SetRange(sensorRange Range) bool {
	d.dataFormat.sensorRange = sensorRange & 0x03
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_DATA_FORMAT, []byte{d.dataFormat.toByte()})
	return true
}

// convertToIS adjusts the raw values from the adxl345 with the range configuration
func (d *dataFormat) convertToIS(rawValue int16) int16 {
	switch d.sensorRange {
	case RANGE_2G:
		return rawValue * 4 // rawValue * 2 * 1000 / 512
	case RANGE_4G:
		return rawValue * 8 // rawValue * 4 * 1000 / 512
	case RANGE_8G:
		return rawValue * 16 // rawValue * 8 * 1000 / 512
	case RANGE_16G:
		return rawValue * 32 // rawValue * 16 * 1000 / 512
	default:
		return 0
	}
}

// toByte returns a byte from the powerCtl configuration
func (p *powerCtl) toByte() (bits uint8) {
	bits = 0x00
	bits = bits | (p.link << 5)
	bits = bits | (p.autoSleep << 4)
	bits = bits | (p.measure << 3)
	bits = bits | (p.sleep << 2)
	bits = bits | p.wakeUp

	return bits
}

// toByte returns a byte from the dataFormat configuration
func (d *dataFormat) toByte() (bits uint8) {
	bits = 0x00
	bits = bits | (d.selfTest << 7)
	bits = bits | (d.spi << 6)
	bits = bits | (d.intInvert << 5)
	bits = bits | (d.fullRes << 3)
	bits = bits | (d.justify << 2)
	bits = bits | uint8(d.sensorRange)

	return bits
}

// toByte returns a byte from the bwRate configuration
func (b *bwRate) toByte() (bits uint8) {
	bits = 0x00
	bits = bits | (b.lowPower << 4)
	bits = bits | uint8(b.rate)

	return bits
}

// readInt converts two bytes to int16
func readIntLE(msb byte, lsb byte) int16 {
	return int16(uint16(msb) | uint16(lsb)<<8)
}
