// Package mpu6886 provides a driver for the MPU6886 accelerometer and gyroscope
// made by InvenSense.
//
// Datasheet:
// https://m5stack.oss-cn-shenzhen.aliyuncs.com/resource/docs/datasheet/core/MPU-6886-000193%2Bv1.1_GHIC_en.pdf
package mpu6886 // import "tinygo.org/x/drivers/mpu6886"

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

const WhoAmI = 0x19

var errNotConnected = errors.New("mpu6886: failed to communicate with a sensor")

// Device wraps an I2C connection to a MPU6886 device.
type Device struct {
	bus     drivers.I2C
	Address uint16
	aRange  uint8
	gRange  uint8
}

// Config contains settings for filtering, sampling, and modes of operation
type Config struct {
	AccelRange uint8
	GyroRange  uint8
}

// New creates a new MPU6886 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) *Device {
	return &Device{bus: bus, Address: DefaultAddress}
}

// Connected returns whether a MPU6886 has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {
	data := []byte{0}
	d.bus.Tx(d.Address, []byte{WHO_AM_I}, data)
	return data[0] == WhoAmI
}

// Configure sets up the device for communication.
func (d *Device) Configure(config Config) (err error) {
	if config.AccelRange < 4 {
		d.aRange = config.AccelRange
	}
	if config.GyroRange < 4 {
		d.gRange = config.GyroRange
	}

	if !d.Connected() {
		return errNotConnected
	}
	// This initialization sequence is borrowed from Arduino M5Stack library
	// Zero register
	if err = d.bus.Tx(d.Address, []byte{PWR_MGMT_1, 0x00}, nil); err != nil {
		return
	}
	time.Sleep(10 * time.Millisecond)
	// Set DEVICE_RESET bit
	if err = d.bus.Tx(d.Address, []byte{PWR_MGMT_1, 0x80}, nil); err != nil {
		return
	}
	time.Sleep(10 * time.Millisecond)
	// Set CLKSEL to 1 - Auto selects the best available clock source
	if err = d.bus.Tx(d.Address, []byte{PWR_MGMT_1, 0x01}, nil); err != nil {
		return
	}
	time.Sleep(10 * time.Millisecond)
	// Set ACCEL_FS_SEL
	if err = d.bus.Tx(d.Address, []byte{ACCEL_CONFIG, d.aRange << 3}, nil); err != nil {
		return
	}
	time.Sleep(time.Millisecond)
	// Set FS_SEL
	if err = d.bus.Tx(d.Address, []byte{GYRO_CONFIG, d.gRange << 3}, nil); err != nil {
		return
	}
	time.Sleep(time.Millisecond)
	// default: 0x80, set DLPF_CFG to 001 (Low Pass Filter)
	if err = d.bus.Tx(d.Address, []byte{CONFIG, 0x01}, nil); err != nil {
		return
	}
	time.Sleep(time.Millisecond)
	// Set sample rate divisor, sample rate is ~ 170 Hz
	if err = d.bus.Tx(d.Address, []byte{SMPLRT_DIV, 0x05}, nil); err != nil {
		return
	}
	time.Sleep(time.Millisecond)
	// Set Interupt pin
	if err = d.bus.Tx(d.Address, []byte{INT_PIN_CFG, 0x22}, nil); err != nil {
		return
	}
	time.Sleep(time.Millisecond)
	// Enable DATA_RDY_INT_EN
	if err = d.bus.Tx(d.Address, []byte{INT_ENABLE, 0x01}, nil); err != nil {
		return
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}

// ReadTemperature returns the temperature in Celsius millidegrees (°C/1000).
func (d *Device) ReadTemperature() (t int32, err error) {
	data := make([]byte, 2)
	if err = d.bus.Tx(d.Address, []byte{TEMP_OUT_H}, data); err != nil {
		return
	}
	rawTemperature := int32(int16((uint16(data[0]) << 8) | uint16(data[1])))
	// The formula to convert to degrre of Celsius is
	//     T_C = T_raw / 326.8 + 25.0
	// This formula should not overflow
	t = rawTemperature*10000/3268 + 25000
	return
}

// ReadAcceleration reads the current acceleration from the device and returns
// it in µg (micro-gravity). When one of the axes is pointing straight to Earth
// and the sensor is not moving the returned value will be around 1000000 or
// -1000000.
func (d *Device) ReadAcceleration() (x int32, y int32, z int32, err error) {
	data := make([]byte, 6)
	if err = d.bus.Tx(d.Address, []byte{ACCEL_XOUT_H}, data); err != nil {
		return
	}
	// Now do two things:
	// 1. merge the two values to a 16-bit number (and cast to a 32-bit integer)
	// 2. scale the value to bring it in the -1000000..1000000 range.
	//    This is done with a trick. What we do here is essentially multiply by
	//    1000000 and divide by 16384 to get the original scale, but to avoid
	//    overflow we do it at 1/64 of the value:
	//      1000000 / 64 = 15625
	//      16384   / 64 = 256
	divider := int32(1)
	switch d.aRange {
	case AFS_RANGE_2_G:
		divider = 256
	case AFS_RANGE_4_G:
		divider = 128
	case AFS_RANGE_8_G:
		divider = 64
	case AFS_RANGE_16_G:
		divider = 32
	}
	x = int32(int16((uint16(data[0])<<8)|uint16(data[1]))) * 15625 / divider
	y = int32(int16((uint16(data[2])<<8)|uint16(data[3]))) * 15625 / divider
	z = int32(int16((uint16(data[4])<<8)|uint16(data[5]))) * 15625 / divider
	return
}

// ReadRotation reads the current rotation from the device and returns it in
// µ°/s (micro-degrees/sec). This means that if you were to do a complete
// rotation along one axis and while doing so integrate all values over time,
// you would get a value close to 360000000.
func (d *Device) ReadRotation() (x int32, y int32, z int32, err error) {
	data := make([]byte, 6)
	if err = d.bus.Tx(d.Address, []byte{GYRO_XOUT_H}, data); err != nil {
		return
	}
	// First the value is converted from a pair of bytes to a signed 16-bit
	// value and then to a signed 32-bit value to avoid integer overflow.
	// Then the value is scaled to µ°/s (micro-degrees per second).
	// This is done in the following steps:
	// 1. Multiply by 250 * 1000_000
	// 2. Divide by 32768
	// The following calculation (x * 15625 / 2048 * 1000) is essentially the
	// same but avoids overflow. First both operations are divided by 16 leading
	// to multiply by 15625000 and divide by 2048, and then part of the multiply
	// is done after the divide instead of before.
	divider := int32(1)
	switch d.gRange {
	case GFS_RANGE_250:
		divider = 2048
	case GFS_RANGE_500:
		divider = 1024
	case GFS_RANGE_1000:
		divider = 512
	case GFS_RANGE_2000:
		divider = 256
	}
	x = int32(int16((uint16(data[0])<<8)|uint16(data[1]))) * 15625 / divider * 1000
	y = int32(int16((uint16(data[2])<<8)|uint16(data[3]))) * 15625 / divider * 1000
	z = int32(int16((uint16(data[4])<<8)|uint16(data[5]))) * 15625 / divider * 1000
	return
}
