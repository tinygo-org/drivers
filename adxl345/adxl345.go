// Package  provides a driver for the digital accelerometer ADXL345
//
// Datasheet EN: http://www.analog.com/media/en/technical-documentation/data-sheets/ADXL345.pdf
// Datasheet JP: http://www.analog.com/media/jp/technical-documentation/data-sheets/ADXL345_jp.pdf
package adxl345

import (
	"machine"
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

// Device wraps an I2C connection to a BMP180 device.
type Device struct {
	bus              machine.I2C
	Address          uint16
	powerCtl         powerCtl
	dataFormat       dataFormat
	bwRate           bwRate
	x, y, z          int32
	rawX, rawY, rawZ int32
}

// New creates a new BMP180 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus machine.I2C, address uint16) Device {
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
		Address: address,
	}
}

// Configure sets up the device for communication
func (d *Device) Configure() {
	d.bus.WriteRegister(uint8(d.Address), REG_BW_RATE, []byte{d.bwRate.toByte()})
	d.bus.WriteRegister(uint8(d.Address), REG_POWER_CTL, []byte{d.powerCtl.toByte()})
	d.bus.WriteRegister(uint8(d.Address), REG_DATA_FORMAT, []byte{d.dataFormat.toByte()})
}

// Halt stops the sensor, values will not updated
func (d *Device) Halt() {
	d.powerCtl.measure = 0
	d.bus.WriteRegister(uint8(d.Address), REG_POWER_CTL, []byte{d.powerCtl.toByte()})
}

// Restart makes reading the sensor working again after a halt
func (d *Device) Restart() {
	d.powerCtl.measure = 1
	d.bus.WriteRegister(uint8(d.Address), REG_POWER_CTL, []byte{d.powerCtl.toByte()})
}

// Acceleration returns the adjusted x, y and z axis in µG
func (d *Device) Acceleration() (x int32, y int32, z int32) {
	return d.x, d.y, d.z
}

// XYZ returns the raw x, y and z axis from the adxl345
func (d *Device) RawXYZ() (x int32, y int32, z int32) {
	return d.rawX, d.rawY, d.rawZ
}

// Update reads the sensor values and stores them in a buffer
func (d *Device) Update() {
	data := []byte{0, 0, 0, 0, 0, 0}
	d.bus.ReadRegister(uint8(d.Address), REG_DATAX0, data)

	d.rawX = readIntLE(data[0], data[1])
	d.rawY = readIntLE(data[2], data[3])
	d.rawZ = readIntLE(data[4], data[5])

	d.x = d.dataFormat.convertToIS(d.rawX)
	d.y = d.dataFormat.convertToIS(d.rawY)
	d.z = d.dataFormat.convertToIS(d.rawZ)
}

// SetRate change the current rate of the sensor
func (d *Device) UseLowPower(power bool) {
	if power {
		d.bwRate.lowPower = 1
	} else {
		d.bwRate.lowPower = 0
	}
	d.bus.WriteRegister(uint8(d.Address), REG_BW_RATE, []byte{d.bwRate.toByte()})
}

// SetRate change the current rate of the sensor
func (d *Device) SetRate(rate Rate) bool {
	d.bwRate.rate = rate & 0x0F
	d.bus.WriteRegister(uint8(d.Address), REG_BW_RATE, []byte{d.bwRate.toByte()})
	return true
}

// SetRange change the current range of the sensor
func (d *Device) SetRange(sensorRange Range) bool {
	d.dataFormat.sensorRange = sensorRange & 0x03
	d.bus.WriteRegister(uint8(d.Address), REG_DATA_FORMAT, []byte{d.dataFormat.toByte()})
	return true
}

// convertToIS adjusts the raw values from the adxl345 with the range configuration
func (d *dataFormat) convertToIS(rawValue int32) int32 {
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
func readIntLE(msb byte, lsb byte) int32 {
	return int32(uint16(msb) | uint16(lsb)<<8)
}
