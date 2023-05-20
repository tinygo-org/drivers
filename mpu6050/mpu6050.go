// Package mpu6050 provides a driver for the MPU6050 accelerometer and gyroscope
// made by InvenSense.
//
// Datasheets:
// https://store.invensense.com/datasheets/invensense/MPU-6050_DataSheet_V3%204.pdf
// https://www.invensense.com/wp-content/uploads/2015/02/MPU-6000-Register-Map1.pdf
package mpu6050 // import "tinygo.org/x/drivers/mpu6050"

import (
	"encoding/binary"
	"errors"

	"tinygo.org/x/drivers"
)

const DefaultAddress = 0x68

// RangeAccel defines the range of the accelerometer.
// Allowed values are 2, 4, 8 and 16 with the unit g (gravity).
type RangeAccel uint8

// RangeGyro defines the range of the gyroscope.
// Allowed values are 250, 500, 1000 and 2000 with the unit °/s (degree per second).
type RangeGyro uint8

var (
	errInvalidRangeAccel = errors.New("mpu6050: invalid range for accelerometer")
	errInvalidRangeGyro  = errors.New("mpu6050: invalid range for gyroscope")
)

type Config struct {
	// Use ACCEL_RANGE_2 through ACCEL_RANGE_16.
	AccelRange RangeAccel
	// Use GYRO_RANGE_250 through GYRO_RANGE_2000
	GyroRange   RangeGyro
	sampleRatio byte // TODO(soypat): expose these as configurable.
	clkSel      byte
}

// Device contains MPU board abstraction for usage
type Device struct {
	conn   drivers.I2C
	aRange int32 //Gyroscope FSR acording to SetAccelRange input
	gRange int32 //Gyroscope FSR acording to SetGyroRange input
	// data contains the accelerometer, gyroscope and temperature data read
	// in the last call via the Update method. The data is stored as seven 16bit unsigned
	// integers in big endian format:
	//
	//	| ax | ay | az | temp | gx | gy | gz |
	data    [14]byte
	address byte
}

// New instantiates and initializes a MPU6050 struct without writing/reading
// i2c bus. Typical I2C MPU6050 address is 0x68.
func New(bus drivers.I2C, addr uint16) *Device {
	p := &Device{}
	p.address = uint8(addr)
	p.conn = bus
	return p
}

// Init configures the necessary registers for using the
// MPU6050. It sets the range of both the accelerometer
// and the gyroscope, the sample rate, the clock source
// and wakes up the peripheral.
func (p *Device) Configure(data Config) (err error) {
	if err = p.Sleep(false); err != nil {
		return err
	}
	if err = p.setClockSource(data.clkSel); err != nil {
		return err
	}
	if err = p.setSampleRate(data.sampleRatio); err != nil {
		return err
	}
	if err = p.setRangeGyro(data.GyroRange); err != nil {
		return err
	}
	if err = p.setRangeAccel(data.AccelRange); err != nil {
		return err
	}
	return nil
}

func (d Device) Connected() bool {
	data := []byte{0}
	d.read(_WHO_AM_I, data)
	return data[0] == 0x68
}

// Update fetches the latest data from the MPU6050
func (p *Device) Update() (err error) {
	if err = p.read(_ACCEL_XOUT_H, p.data[:]); err != nil {
		return err
	}
	return nil
}

// Acceleration returns last read acceleration in µg (micro-gravity).
// When one of the axes is pointing straight to Earth and the sensor is not
// moving the returned value will be around 1000000 or -1000000.
func (d *Device) Acceleration() (ax, ay, az int32) {
	const accelOffset = 0
	ax = int32(convertWord(d.data[accelOffset+0:])) * 15625 / 512 * d.aRange
	ay = int32(convertWord(d.data[accelOffset+2:])) * 15625 / 512 * d.aRange
	az = int32(convertWord(d.data[accelOffset+4:])) * 15625 / 512 * d.aRange
	return ax, ay, az
}

// AngularVelocity reads the current angular velocity from the device and returns it in
// µ°/rad (micro-radians/sec). This means that if you were to do a complete
// rotation along one axis and while doing so integrate all values over time,
// you would get a value close to 6.3 radians (360 degrees).
func (d *Device) AngularVelocity() (gx, gy, gz int32) {
	const angvelOffset = 8
	_ = d.data[angvelOffset+5] // This line fails to compile if RawData is too short.
	gx = int32(convertWord(d.data[angvelOffset+0:])) * 4363 / 8192 * d.gRange
	gy = int32(convertWord(d.data[angvelOffset+2:])) * 4363 / 8192 * d.gRange
	gz = int32(convertWord(d.data[angvelOffset+4:])) * 4363 / 8192 * d.gRange
	return gx, gy, gz
}

// Temperature returns the temperature of the device in milli-centigrade.
func (d *Device) Temperature() (Celsius int32) {
	const tempOffset = 6
	return 1506*int32(convertWord(d.data[tempOffset:]))/512 + 37*1000
}

func convertWord(buf []byte) int16 {
	return int16(binary.BigEndian.Uint16(buf))
}

// setSampleRate sets the sample rate for the FIFO,
// register ouput and DMP. The sample rate is determined
// by:
//
//	SR = Gyroscope Output Rate / (1 + srDiv)
//
// The Gyroscope Output Rate is 8kHz when the DLPF is
// disabled and 1kHz otherwise. The maximum sample rate
// for the accelerometer is 1kHz, if a higher sample rate
// is chosen, the same accelerometer sample will be output.
func (p *Device) setSampleRate(srDiv byte) (err error) {
	// setSampleRate
	var sr [1]byte
	sr[0] = srDiv
	if err = p.write8(_SMPRT_DIV, sr[0]); err != nil {
		return err
	}
	return nil
}

// setClockSource configures the source of the clock
// for the peripheral.
func (p *Device) setClockSource(clkSel byte) (err error) {
	return p.writeMasked(_PWR_MGMT_1, _CLK_SEL_MASK, clkSel)
}

// setRangeGyro configures the full scale range of the gyroscope.
// It has four possible values +- 250°/s, 500°/s, 1000°/s, 2000°/s.
func (p *Device) setRangeGyro(gyroRange RangeGyro) (err error) {
	switch gyroRange {
	case RangeGyro250:
		p.gRange = 250
	case RangeGyro500:
		p.gRange = 500
	case RangeGyro1000:
		p.gRange = 1000
	case RangeGyro2000:
		p.gRange = 2000
	default:
		return errInvalidRangeGyro
	}
	return p.writeMasked(_GYRO_CONFIG, _G_FS_SEL, uint8(gyroRange)<<_G_FS_SHIFT)
}

// setRangeAccel configures the full scale range of the accelerometer.
// It has four possible values +- 2g, 4g, 8g, 16g.
// The function takes values of accRange from 0-3 where 0 means the
// lowest FSR (2g) and 3 is the highest FSR (16g)
func (p *Device) setRangeAccel(accRange RangeAccel) (err error) {
	switch accRange {
	case RangeAccel2:
		p.aRange = 2
	case RangeAccel4:
		p.aRange = 4
	case RangeAccel8:
		p.aRange = 8
	case RangeAccel16:
		p.aRange = 16
	default:
		return errInvalidRangeAccel
	}
	return p.writeMasked(_ACCEL_CONFIG, _AFS_SEL, uint8(accRange)<<_AFS_SHIFT)
}

// Sleep sets the sleep bit on the power managment 1 field.
// When the recieved bool is true, it sets the bit to 1 thus putting
// the peripheral in sleep mode.
// When false is recieved the bit is set to 0 and the peripheral wakes up.
func (p *Device) Sleep(sleepEnabled bool) (err error) {
	return p.writeMasked(_PWR_MGMT_1, _SLEEP_MASK, b2u8(sleepEnabled)<<_SLEEP_SHIFT)
}

func (d *Device) writeMasked(reg byte, mask byte, value byte) error {
	var b [1]byte
	if err := d.read(reg, b[:]); err != nil {
		return err
	}
	b[0] = (b[0] &^ mask) | value&mask
	return d.write8(reg, b[0])
}

func b2u8(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func DefaultConfig() Config {
	return Config{
		AccelRange:  RangeAccel16,
		GyroRange:   RangeGyro2000,
		sampleRatio: 0, // TODO add const values.
		clkSel:      0,
	}
}
