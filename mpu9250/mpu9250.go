// Package mpu9250 provides a driver for the MPU9250 accelerometer, gyroscope and
// magnetometer made by InvenSense.
//
// Datasheets:
// https://www.invensense.com/wp-content/uploads/2015/02/PS-MPU-9250A-01-v1.1.pdf
// https://www.invensense.com/wp-content/uploads/2015/02/RM-MPU-9250A-00-v1.6.pdf
package mpu9250

import (
	"time"

	"github.com/aykevl/tinygo/src/machine"
)

type Data struct {
	accel [3]int32
	gyro  [3]int32
	mag   [3]int32
}

// Device wraps an I2C connection to a MPU9250 device.
type Device struct {
	bus        machine.I2C
	accelRange int32
	gyroRange  int32
	magAdjust  [3]int32
	data       Data
}

// New creates a new MPU9250 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus machine.I2C) Device {
	return Device{
		bus: bus,
	}
}

// Connected returns whether a MPU9250 has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {
	data := []byte{0}
	d.bus.ReadRegister(MPU9250_Address, WHO_AM_I, data)
	return data[0] == WHO_AM_I_RESPONSE
}

// Configure sets up the device for communication.
func (d *Device) Configure() {
	d.SetAccelRange(ACCEL_RANGE_16G)
	d.SetGyroRange(GYRO_RANGE_2000DPS)
	d.configureMag(MAG_MODE_CONTINUOUS_8HZ)
}

// SetAccelRange sets the accelerometer range (2G, 4G, 8G or 16G)
func (d *Device) SetAccelRange(accelRange uint8) {
	switch accelRange {
	case ACCEL_RANGE_2G:
		d.writeRegMPU(ACCEL_CONFIG, ACCEL_FS_SEL_2G)
		d.accelRange = 2
		break
	case ACCEL_RANGE_4G:
		d.writeRegMPU(ACCEL_CONFIG, ACCEL_FS_SEL_4G)
		d.accelRange = 4
		break
	case ACCEL_RANGE_8G:
		d.writeRegMPU(ACCEL_CONFIG, ACCEL_FS_SEL_8G)
		d.accelRange = 8
		break
	case ACCEL_RANGE_16G:
		d.writeRegMPU(ACCEL_CONFIG, ACCEL_FS_SEL_16G)
		d.accelRange = 16
		break
	}
}

// magReadAdjustValues reads and stores the adjust values for the magnetometer
func (d *Device) magReadAdjustValues() {
	d.MagSetMode(MAG_MODE_POWERDOWN)
	d.MagSetMode(MAG_MODE_FUSEROM)
	data := make([]byte, 3)
	d.bus.ReadRegister(AK8963_Address, AK_ASAX, data)
	d.magAdjust[0] = 1000 + int32(float32(1000*int16(uint16(data[0]))-128)/256.0)
	d.magAdjust[1] = 1000 + int32(float32(1000*int16(uint16(data[1]))-128)/256.0)
	d.magAdjust[2] = 1000 + int32(float32(1000*int16(uint16(data[2]))-128)/256.0)
}

// configureMag enables the magnetometer and configures it
func (d *Device) configureMag(mode uint8) {
	d.writeRegMPU(INT_PIN_CFG, 0x02) // bypass enable
	time.Sleep(10 * time.Millisecond)
	d.magReadAdjustValues()
	d.MagSetMode(MAG_MODE_POWERDOWN)
	time.Sleep(10 * time.Millisecond)
	d.MagSetMode(mode)
	time.Sleep(10 * time.Millisecond)
}

// MagSetMode sets the mode (8Hz, 100Hz) of the magnetometer
func (d *Device) MagSetMode(mode uint8) {
	d.writeRegAK(AK_CNTL1, mode)
	time.Sleep(10 * time.Millisecond)
}

// MagHorizDirection returns the horizontal direction of the
// magnetometer in millidegrees
func (d *Device) MagHorizDirection() int32 {
	x, y, _ := d.Magnetometer()
	return int32(1000 * (atan(float32(x)/float32(y)) * 180.0 / PI))
}

// MagUpdate updates the buffer's values of the magnetometer
func (d *Device) MagUpdate() {
	x, y, z := d.RawMagnetometer()
	d.data.mag[0] = int32(x) * d.magAdjust[0]
	d.data.mag[1] = int32(y) * d.magAdjust[1]
	d.data.mag[2] = int32(z) * d.magAdjust[2]
}

// Magnetometer returns the (x,y,z) values of the AK8963 in µT
func (d *Device) Magnetometer() (x int32, y int32, z int32) {
	return d.data.mag[0], d.data.mag[1], d.data.mag[2]
}

// AccelUpdate updates the buffer's values of the accelerometer
func (d *Device) AccelUpdate() {
	x, y, z := d.RawAcceleration()
	d.data.accel[0] = (1000 * int32(x) * d.accelRange) / 32768
	d.data.accel[1] = (1000 * int32(y) * d.accelRange) / 32768
	d.data.accel[2] = (1000 * int32(z) * d.accelRange) / 32768
}

// Acceleration returns the (x,y,z) values of the accelerometer in mG
func (d *Device) Acceleration() (x int32, y int32, z int32) {
	return d.data.accel[0], d.data.accel[1], d.data.accel[2]
}

// SetGyroRange sets the range (DPS) of the gyroscope
func (d *Device) SetGyroRange(mode uint8) {
	switch mode {
	case GYRO_RANGE_250DPS:
		d.writeRegMPU(GYRO_CONFIG, GYRO_FS_SEL_250DPS)
		d.gyroRange = 250
		break
	case GYRO_RANGE_500DPS:
		d.writeRegMPU(GYRO_CONFIG, GYRO_FS_SEL_500DPS)
		d.gyroRange = 500
		break
	case GYRO_RANGE_1000DPS:
		d.writeRegMPU(GYRO_CONFIG, GYRO_FS_SEL_1000DPS)
		d.gyroRange = 1000
		break
	case GYRO_RANGE_2000DPS:
		d.writeRegMPU(GYRO_CONFIG, GYRO_FS_SEL_2000DPS)
		d.gyroRange = 2000
		break
	}
}

// GyroUpdate updates the buffer's values of the gyroscope
func (d *Device) GyroUpdate() {
	x, y, z := d.RawRotation()
	d.data.gyro[0] = (1000 * int32(x) * d.gyroRange) / 32768
	d.data.gyro[1] = (1000 * int32(y) * d.gyroRange) / 32768
	d.data.gyro[2] = (1000 * int32(z) * d.gyroRange) / 32768

}

// Rotation returns the (x,y,z) values of the gyroscope in millidegrees per second
func (d *Device) Rotation() (x int32, y int32, z int32) {
	return d.data.gyro[0], d.data.gyro[1], d.data.gyro[2]
}

// RawAcceleration returns the (x,y,z) raw values of the accelerometer sensor
func (d *Device) RawAcceleration() (x int16, y int16, z int16) {
	data := make([]byte, 6)
	d.bus.ReadRegister(MPU9250_Address, ACCEL_XOUT_H, data)
	x = readInt(data[0], data[1])
	y = readInt(data[2], data[3])
	z = readInt(data[4], data[5])
	return
}

// RawRotation returns the (x,y,z) raw values of the gyroscope sensor
func (d *Device) RawRotation() (x int16, y int16, z int16) {
	data := make([]byte, 6)
	d.bus.ReadRegister(MPU9250_Address, GYRO_XOUT_H, data)
	x = readInt(data[0], data[1])
	y = readInt(data[2], data[3])
	z = readInt(data[4], data[5])
	return
}

// RawMagnetometer returns the (x,y,z) raw values of the magnetometer sensor
func (d *Device) RawMagnetometer() (x int16, y int16, z int16) {
	data := make([]byte, 7)
	d.bus.ReadRegister(AK8963_Address, AK_HXL, data)
	x = readInt(data[1], data[0])
	y = readInt(data[3], data[2])
	z = readInt(data[5], data[4])
	return
}

// writeMPU writes the byte to the specified register of the MPU9250
func (d *Device) writeRegMPU(reg byte, data byte) error {
	return d.bus.WriteRegister(MPU9250_Address, reg, []byte{data})
}

// writeMPU writes the byte to the specified register of the AK8963
func (d *Device) writeRegAK(reg byte, data byte) error {
	return d.bus.WriteRegister(AK8963_Address, reg, []byte{data})
}

// readInt converts two bytes to int16
func readInt(msb byte, lsb byte) int16 {
	return int16(uint16(msb)<<8 | uint16(lsb))
}

// readUint converts two bytes to uint16
func readUint(msb byte, lsb byte) uint16 {
	return (uint16(msb) << 8) | uint16(lsb)
}

// TODO: remove when math import is fixed
// math.atan function copied from:
// https://github.com/golang/go/blob/master/src/math/atan2.go
// xatan evaluates a series valid in the range [0, 0.66].
func xatan(x float32) float32 {
	const (
		P0 = -8.750608600031904122785e-01
		P1 = -1.615753718733365076637e+01
		P2 = -7.500855792314704667340e+01
		P3 = -1.228866684490136173410e+02
		P4 = -6.485021904942025371773e+01
		Q0 = +2.485846490142306297962e+01
		Q1 = +1.650270098316988542046e+02
		Q2 = +4.328810604912902668951e+02
		Q3 = +4.853903996359136964868e+02
		Q4 = +1.945506571482613964425e+02
	)
	z := x * x
	z = z * ((((P0*z+P1)*z+P2)*z+P3)*z + P4) / (((((z+Q0)*z+Q1)*z+Q2)*z+Q3)*z + Q4)
	z = x*z + x
	return z
}

// satan reduces its argument (known to be positive)
// to the range [0, 0.66] and calls xatan.
func satan(x float32) float32 {
	const (
		Morebits = 6.123233995736765886130e-17 // pi/2 = PIO2 + Morebits
		Tan3pio8 = 2.41421356237309504880      // tan(3*pi/8)
	)
	if x <= 0.66 {
		return xatan(x)
	}
	if x > Tan3pio8 {
		return PI/2 - xatan(1/x) + Morebits
	}
	return PI/4 + xatan((x-1)/(x+1)) + 0.5*Morebits
}

func atan(x float32) float32 {
	if x == 0 {
		return x
	}
	if x > 0 {
		return satan(x)
	}
	return -satan(-x)
}
