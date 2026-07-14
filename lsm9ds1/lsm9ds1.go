// LSM9DS1, 9 axis Inertial Measurement Unit (IMU)
//
// Datasheet: https://www.st.com/resource/en/datasheet/lsm6ds3.pdf
package lsm9ds1 // import "tinygo.org/x/drivers/lsm9ds1"

import (
	"errors"

	"tinygo.org/x/drivers"
)

type AccelRange uint8
type AccelSampleRate uint8
type AccelBandwidth uint8

type GyroRange uint8
type GyroSampleRate uint8

type MagRange uint8
type MagSampleRate uint8

// Device wraps connection to a LSM9DS1 device.
type Device struct {
	bus             drivers.I2C
	AccelAddress    uint8
	MagAddress      uint8
	accelMultiplier int32
	gyroMultiplier  int32
	magMultiplier   int32
	buf             [7]uint8 // up to 6 bytes for read + 1 byte for the register address
}

// Configuration for LSM9DS1 device.
type Configuration struct {
	AccelRange      AccelRange
	AccelSampleRate AccelSampleRate
	AccelBandWidth  AccelBandwidth
	GyroRange       GyroRange
	GyroSampleRate  GyroSampleRate
	MagRange        MagRange
	MagSampleRate   MagSampleRate
}

var errNotConnected = errors.New("lsm9ds1: failed to communicate with either acel/gyro or magnet sensor")

// New creates a new LSM9DS1 connection. The I2C bus must already be configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) *Device {
	return &Device{
		bus:          bus,
		AccelAddress: ACCEL_ADDRESS,
		MagAddress:   MAG_ADDRESS,
	}
}

// Connected returns whether both sensor on LSM9DS1 has been found.
// It does two "who am I" requests and checks the responses.
// In a rare case of an I2C bus issue, it can also return an error.
// Case of boolean false and error nil means I2C is up,
// but "who am I" responses have unexpected values.
func (d *Device) Connected() bool {
	data, err := d.readBytes(d.AccelAddress, WHO_AM_I, 1)
	if err != nil || data[0] != 0x68 {
		return false
	}
	data, err = d.readBytes(d.MagAddress, WHO_AM_I_M, 1)
	if err != nil || data[0] != 0x3D {
		return false
	}
	return true
}

// ReadAcceleration reads the current acceleration from the device and returns
// it in µg (micro-gravity). When one of the axes is pointing straight to Earth
// and the sensor is not moving the returned value will be around 1000000 or
// -1000000.
func (d *Device) ReadAcceleration() (x, y, z int32, err error) {
	data, err := d.readBytes(d.AccelAddress, OUT_X_L_XL, 6)
	if err != nil {
		return
	}
	x = int32(int16((uint16(data[1])<<8)|uint16(data[0]))) * d.accelMultiplier
	y = int32(int16((uint16(data[3])<<8)|uint16(data[2]))) * d.accelMultiplier
	z = int32(int16((uint16(data[5])<<8)|uint16(data[4]))) * d.accelMultiplier
	return
}

// ReadRotation reads the current rotation from the device and returns it in
// µ°/s (micro-degrees/sec). This means that if you were to do a complete
// rotation along one axis and while doing so integrate all values over time,
// you would get a value close to 360000000.
func (d *Device) ReadRotation() (x, y, z int32, err error) {
	data, err := d.readBytes(d.AccelAddress, OUT_X_L_G, 6)
	if err != nil {
		return
	}
	x = int32(int16((uint16(data[1])<<8)|uint16(data[0]))) * d.gyroMultiplier
	y = int32(int16((uint16(data[3])<<8)|uint16(data[2]))) * d.gyroMultiplier
	z = int32(int16((uint16(data[5])<<8)|uint16(data[4]))) * d.gyroMultiplier
	return
}

// ReadMagneticField reads the current magnetic field from the device and returns
// it in nT (nanotesla). 1 G (gauss) = 100_000 nT (nanotesla).
func (d *Device) ReadMagneticField() (x, y, z int32, err error) {
	data, err := d.readBytes(d.MagAddress, OUT_X_L_M, 6)
	if err != nil {
		return
	}
	x = int32(int16((int16(data[1])<<8)|int16(data[0]))) * d.magMultiplier
	y = int32(int16((int16(data[3])<<8)|int16(data[2]))) * d.magMultiplier
	z = int32(int16((int16(data[5])<<8)|int16(data[4]))) * d.magMultiplier
	return
}

// ReadTemperature returns the temperature in Celsius milli degrees (°C/1000)
func (d *Device) ReadTemperature() (t int32, err error) {
	data, err := d.readBytes(d.AccelAddress, OUT_TEMP_L, 2)
	if err != nil {
		return
	}
	// From "Table 5. Temperature sensor characteristics"
	// temp = value/16 + 25
	t = 25000 + (int32(int16((int16(data[1])<<8)|int16(data[0])))*125)/2
	return
}

// --- end of public methods --------------------------------------------------

// doConfigure is called by public Configure methods after all
// necessary board-specific initialisations are taken care of
func (d *Device) doConfigure(cfg Configuration) (err error) {

	// Verify unit communication
	if !d.Connected() {
		return errNotConnected
	}

	// Multipliers come from "Table 3. Sensor characteristics" of the datasheet * 1000
	switch cfg.AccelRange {
	case ACCEL_2G:
		d.accelMultiplier = 61
	case ACCEL_4G:
		d.accelMultiplier = 122
	case ACCEL_8G:
		d.accelMultiplier = 244
	case ACCEL_16G:
		d.accelMultiplier = 732
	}
	switch cfg.GyroRange {
	case GYRO_250DPS:
		d.gyroMultiplier = 8750
	case GYRO_500DPS:
		d.gyroMultiplier = 17500
	case GYRO_2000DPS:
		d.gyroMultiplier = 70000
	}
	switch cfg.MagRange {
	case MAG_4G:
		d.magMultiplier = 14
	case MAG_8G:
		d.magMultiplier = 29
	case MAG_12G:
		d.magMultiplier = 43
	case MAG_16G:
		d.magMultiplier = 58
	}

	// Configure accelerometer
	// Sample rate & measurement range
	err = d.writeByte(d.AccelAddress, CTRL_REG6_XL, uint8(cfg.AccelSampleRate)<<5|uint8(cfg.AccelRange)<<3)
	if err != nil {
		return
	}

	// Configure gyroscope
	// Sample rate & measurement range
	err = d.writeByte(d.AccelAddress, CTRL_REG1_G, uint8(cfg.GyroSampleRate)<<5|uint8(cfg.GyroRange)<<3)
	if err != nil {
		return
	}

	// Configure magnetometer

	// Temperature compensation enabled
	// High-performance mode XY axis
	// Sample rate
	err = d.writeByte(d.MagAddress, CTRL_REG1_M, 0b10000000|0b01000000|uint8(cfg.MagSampleRate)<<2)
	if err != nil {
		return
	}

	// Measurement range
	err = d.writeByte(d.MagAddress, CTRL_REG2_M, uint8(cfg.MagRange)<<5)
	if err != nil {
		return
	}

	// Continuous-conversion mode
	// https://electronics.stackexchange.com/questions/237397/continuous-conversion-vs-single-conversion-mode
	err = d.writeByte(d.MagAddress, CTRL_REG3_M, 0b00000000)
	if err != nil {
		return
	}

	// High-performance mode Z axis
	err = d.writeByte(d.MagAddress, CTRL_REG4_M, 0b00001000)
	if err != nil {
		return
	}

	return nil
}

func (d *Device) readBytes(addr, reg, size uint8) ([]byte, error) {
	d.buf[0] = reg
	err := d.bus.Tx(uint16(addr), d.buf[0:1], d.buf[1:size+1])
	if err != nil {
		return nil, err
	}
	return d.buf[1 : size+1], nil
}

func (d *Device) writeByte(addr, reg, value uint8) error {
	d.buf[0] = reg
	d.buf[1] = value
	return d.bus.Tx(uint16(addr), d.buf[0:2], nil)
}
