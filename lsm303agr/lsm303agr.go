/*
Package lsm303agr implements a driver for the LSM303AGR,
 a 3 axis accelerometer/magnetic sensor which is included on BBC micro:bits v1.5.

Datasheet: https://www.st.com/resource/en/datasheet/lsm303agr.pdf
*/

package lsm303agr // import "tinygo.org/x/drivers/lsm303agr"

import (
	"machine"
	"math"
)

/* for the LSM303AGR object */
type Device struct {
	bus            machine.I2C
	AccelAddress   uint8
	MagAddress     uint8
	AccelPowerMode uint8
	AccelRange     uint8
	AccelDataRate  uint8
	MagPowerMode   uint8
	MagSystemMode  uint8
	MagDataRate    uint8
}

/* for configuring LSM303AGR */
type Configuration struct {
	AccelPowerMode uint8
	AccelRange     uint8
	AccelDataRate  uint8
	MagPowerMode   uint8
	MagSystemMode  uint8
	MagDataRate    uint8
}

/* create a new LSM303AGR object */
func New(bus machine.I2C) Device {
	return Device{bus: bus, AccelAddress: ACCEL_ADDRESS, MagAddress: MAG_ADDRESS}
}

/* check if LSM303AGR's both sensors are connected */
func (d *Device) Connected() bool {
	data1, data2 := []byte{0}, []byte{0}
	d.bus.ReadRegister(uint8(d.AccelAddress), ACCEL_WHO_AM_I, data1)
	d.bus.ReadRegister(uint8(d.MagAddress), MAG_WHO_AM_I, data2)
	return data1[0] == 0x33 && data2[0] == 0x40
}

/* configure and initialize LSM303AGR */
func (d *Device) Configure(cfg Configuration) {

	if cfg.AccelDataRate != 0 {
		d.AccelDataRate = cfg.AccelDataRate
	} else {
		d.AccelDataRate = ACCEL_DATARATE_100HZ
	}

	if cfg.AccelPowerMode != 0 {
		d.AccelPowerMode = cfg.AccelPowerMode
	} else {
		d.AccelPowerMode = ACCEL_POWER_NORMAL
	}

	if cfg.AccelRange != 0 {
		d.AccelRange = cfg.AccelRange
	} else {
		d.AccelRange = ACCEL_RANGE_2G
	}

	if cfg.MagPowerMode != 0 {
		d.MagPowerMode = cfg.MagPowerMode
	} else {
		d.MagPowerMode = MAG_POWER_NORMAL
	}

	if cfg.MagDataRate != 0 {
		d.MagDataRate = cfg.MagDataRate
	} else {
		d.MagDataRate = MAG_DATARATE_10HZ
	}

	if cfg.MagSystemMode != 0 {
		d.MagSystemMode = cfg.MagSystemMode
	} else {
		d.MagSystemMode = MAG_SYSTEM_CONTINUOUS
	}

	cmd := []byte{0}

	cmd[0] = byte(d.AccelDataRate<<4 | d.AccelPowerMode | 0x07)
	d.bus.WriteRegister(uint8(d.AccelAddress), ACCEL_CTRL_REG1_A, cmd)

	cmd[0] = byte(0x80 | d.AccelRange<<4)
	d.bus.WriteRegister(uint8(d.AccelAddress), ACCEL_CTRL_REG4_A, cmd)

	cmd[0] = byte(0xC0)
	d.bus.WriteRegister(uint8(d.AccelAddress), TEMP_CFG_REG_A, cmd)

	cmd[0] = byte(0x80 | d.MagPowerMode<<4 | d.MagDataRate<<2 | d.MagSystemMode)
	d.bus.WriteRegister(uint8(d.MagAddress), MAG_MR_REG_M, cmd)

}

/* read raw acceleration data (in ug/microgram) from all axis */
func (d *Device) ReadAcceleration() (x int32, y int32, z int32) {

	data1, data2, data3, data4, data5, data6 := []byte{0}, []byte{0}, []byte{0}, []byte{0}, []byte{0}, []byte{0}
	d.bus.ReadRegister(uint8(d.AccelAddress), ACCEL_OUT_X_H_A, data1)
	d.bus.ReadRegister(uint8(d.AccelAddress), ACCEL_OUT_X_L_A, data2)
	d.bus.ReadRegister(uint8(d.AccelAddress), ACCEL_OUT_Y_H_A, data3)
	d.bus.ReadRegister(uint8(d.AccelAddress), ACCEL_OUT_Y_L_A, data4)
	d.bus.ReadRegister(uint8(d.AccelAddress), ACCEL_OUT_Z_H_A, data5)
	d.bus.ReadRegister(uint8(d.AccelAddress), ACCEL_OUT_Z_L_A, data6)

	range_factor := int16(0)
	switch d.AccelRange {
	case ACCEL_RANGE_2G:
		range_factor = 1
	case ACCEL_RANGE_4G:
		range_factor = 2
	case ACCEL_RANGE_8G:
		range_factor = 4
	case ACCEL_RANGE_16G:
		range_factor = 12 // the readings in 16G are a bit off
	}

	x = int32(int16((uint16(data1[0])<<8 | uint16(data2[0]))) >> 4 * range_factor) * 1000
	y = int32(int16((uint16(data3[0])<<8 | uint16(data4[0]))) >> 4 * range_factor) * 1000
	z = int32(int16((uint16(data5[0])<<8 | uint16(data6[0]))) >> 4 * range_factor) * 1000
	return
}

/* read pitch/roll degrees */
func (d *Device) ReadPitchRoll() (pitch int32, roll int32) {

	x, y, z := d.ReadAcceleration()
	xf, yf, zf := float64(x), float64(y), float64(z)
	pitch = int32(math.Round(math.Atan2(yf, math.Sqrt(math.Pow(xf, 2)+math.Pow(zf, 2)))*(180/math.Pi)*100) / 100)
	roll = int32(math.Round(math.Atan2(xf, math.Sqrt(math.Pow(yf, 2)+math.Pow(zf, 2)))*(180/math.Pi)*100) / 100)
	return

}

/* read magnetic field level (in milligauss) from all axis */
func (d *Device) ReadMagneticField() (x int32, y int32, z int32) {

	if d.MagSystemMode == MAG_SYSTEM_SINGLE {
		cmd := []byte{0}
		cmd[0] = byte(0x80 | d.MagPowerMode<<4 | d.MagDataRate<<2 | d.MagSystemMode)
		d.bus.WriteRegister(uint8(d.MagAddress), MAG_MR_REG_M, cmd)
	}

	data1, data2, data3, data4, data5, data6 := []byte{0}, []byte{0}, []byte{0}, []byte{0}, []byte{0}, []byte{0}
	d.bus.ReadRegister(uint8(d.MagAddress), MAG_OUT_X_H_M, data1)
	d.bus.ReadRegister(uint8(d.MagAddress), MAG_OUT_X_L_M, data2)
	d.bus.ReadRegister(uint8(d.MagAddress), MAG_OUT_Y_H_M, data3)
	d.bus.ReadRegister(uint8(d.MagAddress), MAG_OUT_Y_L_M, data4)
	d.bus.ReadRegister(uint8(d.MagAddress), MAG_OUT_Z_H_M, data5)
	d.bus.ReadRegister(uint8(d.MagAddress), MAG_OUT_Z_L_M, data6)

	x = int32(int16((uint16(data1[0])<<8 | uint16(data2[0]))))
	y = int32(int16((uint16(data3[0])<<8 | uint16(data4[0]))))
	z = int32(int16((uint16(data5[0])<<8 | uint16(data6[0]))))
	return
}

/* read compass heading, -179~180 degrees (may not be accurate) */
func (d *Device) ReadCompassHeading() (heading int32) {

	x, y, _ := d.ReadMagneticField()
	xf, yf := float64(x), float64(y)
	heading = int32(float32((180 / math.Pi) * math.Atan2(yf, xf)))
	return
}

/* read temperature offset */
func (d *Device) ReadTemperatureOffset() (t int32) {

	data1, data2 := []byte{0}, []byte{0}
	d.bus.ReadRegister(uint8(d.AccelAddress), OUT_TEMP_H_A, data1)
	d.bus.ReadRegister(uint8(d.AccelAddress), OUT_TEMP_L_A, data2)
	t = int32(int16((uint16(data1[0])<<8 | uint16(data2[0]))) >> 4)
	return
}

/* read temperature in Celsius */
func (d *Device) ReadTemperature() (c int32, e error) {

	t := d.ReadTemperatureOffset()
	c = int32((float32(25) + float32(t)/8) * 1000)
	e = nil
	return
}
