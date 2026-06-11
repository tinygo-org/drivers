// Package bmi270 provides a driver for the BMI270 6-axis inertial measurement unit.
//
// Datasheet: https://www.bosch-sensortec.com/media/boschsensortec/downloads/datasheets/bst-bmi270-ds000.pdf
package bmi270

import (
	_ "embed"
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

//go:embed bmi270-config.bin
var bmi270ConfigData string

const Address = 0x68

type AccelRange uint8

const (
	Accel2G  AccelRange = 0x00
	Accel4G  AccelRange = 0x01
	Accel8G  AccelRange = 0x02
	Accel16G AccelRange = 0x03
)

type GyroRange uint8

const (
	Gyro2000DPS GyroRange = 0x00
	Gyro1000DPS GyroRange = 0x01
	Gyro500DPS  GyroRange = 0x02
	Gyro250DPS  GyroRange = 0x03
	Gyro125DPS  GyroRange = 0x04
)

type Device struct {
	bus        drivers.I2C
	address    uint8
	accelRange AccelRange
	gyroRange  GyroRange
	wbuf       [17]byte
	rbuf       [6]byte
}

type Config struct {
	AccelRange AccelRange
	GyroRange  GyroRange
}

func DefaultConfig() Config {
	return Config{
		AccelRange: Accel2G,
		GyroRange:  Gyro2000DPS,
	}
}

var (
	errNotConnected = errors.New("bmi270: not connected")
	errInitFailed   = errors.New("bmi270: initialization failed")
)

func NewI2C(bus drivers.I2C, address uint8) *Device {
	return &Device{
		bus:     bus,
		address: address,
	}
}

func (d *Device) Connected() bool {
	val, err := d.read1(reg_CHIP_ID)
	return err == nil && val == chipIDBMI270
}

func (d *Device) Configure(config Config) error {
	if config.AccelRange != 0 {
		d.accelRange = config.AccelRange
	} else {
		d.accelRange = Accel2G
	}

	if config.GyroRange != 0 {
		d.gyroRange = config.GyroRange
	} else {
		d.gyroRange = Gyro2000DPS
	}

	if !d.Connected() {
		return errNotConnected
	}

	if err := d.write1(reg_CMD, 0xB6); err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)

	if err := d.write1(reg_PWR_CONF, 0x00); err != nil {
		return err
	}
	time.Sleep(1 * time.Millisecond)

	if err := d.write1(reg_INIT_CTRL, 0x00); err != nil {
		return err
	}
	time.Sleep(1 * time.Millisecond)

	configBytes := []byte(bmi270ConfigData)
	chunkSize := 16
	for i := 0; i < len(configBytes); i += chunkSize {
		end := i + chunkSize
		if end > len(configBytes) {
			end = len(configBytes)
		}

		wordAddr := uint16(i / 2)
		addrLow := byte(wordAddr & 0x0F)
		addrHigh := byte(wordAddr >> 4)
		d.wbuf[0] = reg_INIT_ADDR_0
		d.wbuf[1] = addrLow
		d.wbuf[2] = addrHigh
		if err := d.bus.Tx(uint16(d.address), d.wbuf[:3], nil); err != nil {
			return err
		}

		chunk := configBytes[i:end]
		d.wbuf[0] = reg_INIT_DATA
		copy(d.wbuf[1:], chunk)
		if err := d.bus.Tx(uint16(d.address), d.wbuf[:1+len(chunk)], nil); err != nil {
			return err
		}
	}

	if err := d.write1(reg_INIT_CTRL, 0x01); err != nil {
		return err
	}
	time.Sleep(200 * time.Millisecond)

	start := time.Now()
	for {
		status, err := d.read1(reg_INTERNAL_STATUS)
		if err != nil {
			return err
		}
		if status == 0x01 {
			break
		}
		if time.Since(start) >= 500*time.Millisecond {
			return errInitFailed
		}
		time.Sleep(50 * time.Millisecond)
	}

	if err := d.write1(reg_ACC_CONF, 0xA8); err != nil {
		return err
	}

	var rangeVal byte
	switch d.accelRange {
	case Accel2G:
		rangeVal = 0x00
	case Accel4G:
		rangeVal = 0x01
	case Accel8G:
		rangeVal = 0x02
	case Accel16G:
		rangeVal = 0x03
	default:
		rangeVal = 0x00
	}
	if err := d.write1(reg_ACC_RANGE, rangeVal); err != nil {
		return err
	}

	if err := d.write1(reg_GYR_CONF, 0xA8); err != nil {
		return err
	}

	var gyroRangeVal byte
	switch d.gyroRange {
	case Gyro2000DPS:
		gyroRangeVal = 0x00
	case Gyro1000DPS:
		gyroRangeVal = 0x01
	case Gyro500DPS:
		gyroRangeVal = 0x02
	case Gyro250DPS:
		gyroRangeVal = 0x03
	case Gyro125DPS:
		gyroRangeVal = 0x04
	default:
		gyroRangeVal = 0x00
	}
	if err := d.write1(reg_GYR_RANGE, gyroRangeVal); err != nil {
		return err
	}

	if err := d.write1(reg_PWR_CTRL, 0x06); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)

	return nil
}

// ReadAcceleration returns the acceleration in µg (micro-gravity).
// When one of the axes is pointing straight down and the sensor
// is not moving, the returned value will be around 1000000.
func (d *Device) ReadAcceleration() (x, y, z int32, err error) {
	if err = d.bus.Tx(uint16(d.address), []byte{reg_ACC_DATA}, d.rbuf[:6]); err != nil {
		return 0, 0, 0, err
	}

	rawX := int16(uint16(d.rbuf[1])<<8 | uint16(d.rbuf[0]))
	rawY := int16(uint16(d.rbuf[3])<<8 | uint16(d.rbuf[2]))
	rawZ := int16(uint16(d.rbuf[5])<<8 | uint16(d.rbuf[4]))

	k := int32(61)
	switch d.accelRange {
	case Accel4G:
		k = 122
	case Accel8G:
		k = 244
	case Accel16G:
		k = 488
	}

	x = int32(rawX) * k
	y = int32(rawY) * k
	z = int32(rawZ) * k
	return
}

// ReadRotation returns the angular velocity in µdps (micro-degrees/second).
func (d *Device) ReadRotation() (x, y, z int32, err error) {
	if err = d.bus.Tx(uint16(d.address), []byte{reg_GYR_DATA}, d.rbuf[:6]); err != nil {
		return 0, 0, 0, err
	}

	rawX := int16(uint16(d.rbuf[1])<<8 | uint16(d.rbuf[0]))
	rawY := int16(uint16(d.rbuf[3])<<8 | uint16(d.rbuf[2]))
	rawZ := int16(uint16(d.rbuf[5])<<8 | uint16(d.rbuf[4]))

	k := int32(60976)
	switch d.gyroRange {
	case Gyro1000DPS:
		k = 30488
	case Gyro500DPS:
		k = 15244
	case Gyro250DPS:
		k = 7622
	case Gyro125DPS:
		k = 3811
	}

	x = int32(rawX) * k
	y = int32(rawY) * k
	z = int32(rawZ) * k
	return
}

func (d *Device) read1(register uint8) (uint8, error) {
	d.wbuf[0] = register
	err := d.bus.Tx(uint16(d.address), d.wbuf[:1], d.rbuf[:1])
	return d.rbuf[0], err
}

func (d *Device) write1(register, val uint8) error {
	d.wbuf[0] = register
	d.wbuf[1] = val
	return d.bus.Tx(uint16(d.address), d.wbuf[:2], nil)
}
