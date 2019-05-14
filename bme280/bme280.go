package bme280

import (
	"machine"
)

// calibrationCoefficients reads at startup and stores the calibration coefficients
type calibrationCoefficients struct {
	t1 uint16
	t2 int16
	t3 int16
	p1 uint16
	p2 int16
	p3 int16
	p4 int16
	p5 int16
	p6 int16
	p7 int16
	p8 int16
	p9 int16
	h1 uint8
	h2 int16
	h3 uint8
	h4 int16
	h5 int16
	h6 int8
}

// Device wraps an I2C connection to a BME280 device.
type Device struct {
	bus                     machine.I2C
	Address                 uint16
	calibrationCoefficients calibrationCoefficients
}

// New creates a new BME280 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus machine.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
	}
}

// Configure sets up the device for communication and
// read the calibration coefficientes.
func (d *Device) Configure() {

	var data [24]byte
	err := d.bus.ReadRegister(uint8(d.Address), REG_CALIBRATION, data[:])
	if err != nil {
		return
	}

	var h1 [1]byte
	err = d.bus.ReadRegister(uint8(d.Address), REG_CALIBRATION_H1, h1[:])
	if err != nil {
		return
	}

	var h2lsb [7]byte
	err = d.bus.ReadRegister(uint8(d.Address), REG_CALIBRATION_H2LSB, h2lsb[:])
	if err != nil {
		return
	}

	d.calibrationCoefficients.t1 = readUint(data[0], data[1])
	d.calibrationCoefficients.t2 = readInt(data[2], data[3])
	d.calibrationCoefficients.t3 = readInt(data[4], data[5])
	d.calibrationCoefficients.p1 = readUint(data[6], data[7])
	d.calibrationCoefficients.p2 = readInt(data[8], data[9])
	d.calibrationCoefficients.p3 = readInt(data[10], data[11])
	d.calibrationCoefficients.p4 = readInt(data[12], data[13])
	d.calibrationCoefficients.p5 = readInt(data[14], data[15])
	d.calibrationCoefficients.p6 = readInt(data[16], data[17])
	d.calibrationCoefficients.p7 = readInt(data[18], data[19])
	d.calibrationCoefficients.p8 = readInt(data[20], data[21])
	d.calibrationCoefficients.p9 = readInt(data[22], data[23])

	d.calibrationCoefficients.h1 = h1[0]
	d.calibrationCoefficients.h2 = readInt(h2lsb[0], h2lsb[1])
	d.calibrationCoefficients.h3 = h2lsb[2]
	d.calibrationCoefficients.h6 = int8(h2lsb[6])
	d.calibrationCoefficients.h4 = 0 + (int16(h2lsb[3]) << 4) | (int16(h2lsb[4] & 0x0F))
	d.calibrationCoefficients.h5 = 0 + (int16(h2lsb[5]) << 4) | (int16(h2lsb[4]) >> 4)

	d.bus.WriteRegister(uint8(d.Address), CTRL_MEAS_ADDR, []byte{0xB7})
	d.bus.WriteRegister(uint8(d.Address), CTRL_CONFIG, []byte{0x00})

}

// Connected returns whether a BME280 has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {
	data := []byte{0}
	d.bus.ReadRegister(uint8(d.Address), WHO_AM_I, data)
	return data[0] == CHIP_ID
}

func (d *Device) Reset() {
	d.bus.WriteRegister(uint8(d.Address), CMD_RESET, []byte{0xB6})
}

// ReadTemperature returns the temperature in celsius milli degrees (ºC/10)
func (d *Device) ReadTemperature() (int32, error) {
	data, err := d.readData()
	if err != nil {
		return 0, err
	}

	temp, _ := d.calculateTemp(data)
	return temp, nil
}

func (d *Device) ReadPressure() (int32, error) {
	data, err := d.readData()
	if err != nil {
		return 0, err
	}
	_, tFine := d.calculateTemp(data)
	pressure := d.calculatePressure(data, tFine)
	return pressure, nil
}

func (d *Device) ReadHumidity() (int32, error) {
	data, err := d.readData()
	if err != nil {
		return 0, err
	}
	_, tFine := d.calculateTemp(data)
	humidity := d.calculateHumidity(data, tFine)
	return humidity, nil
}

// readInt converts two bytes to int16
func readInt(msb byte, lsb byte) int16 {
	return int16(readUint(msb, lsb))
}

// readUint converts two bytes to uint16
func readUint(msb byte, lsb byte) uint16 {
	temp := (uint16(msb) << 8) | uint16(lsb)
	return (temp >> 8) | (temp << 8)
}

// readData does a burst read from 0xF7 to 0xF0 according to the datasheet
// resulting in an slice with 8 bytes 0-2 = pressure / 3-5 = temperature / 6-7 = humidity
func (d *Device) readData() (data [8]byte, err error) {
	err = d.bus.ReadRegister(uint8(d.Address), REG_PRESSURE, data[:])
	if err != nil {
		println(err)
		return
	}
	return
}

// calculateTemp uses the data slice and applies calibrations values on it to convert the value to an useful integer
// it also calculates the variable tFine which is used by the pressure calculation
func (d *Device) calculateTemp(data [8]byte) (T int32, tFine int32) {

	rawTemp := int32((((uint32(data[3]) << 8) | uint32(data[4])) << 8) | uint32(data[5]))
	rawTemp = rawTemp >> 4

	var1 := (((rawTemp >> 3) - (int32(d.calibrationCoefficients.t1) << 1)) * int32(d.calibrationCoefficients.t2)) >> 11
	var2 := (((((rawTemp >> 4) - int32(d.calibrationCoefficients.t1)) * ((rawTemp >> 4) - int32(d.calibrationCoefficients.t1))) >> 12) * int32(d.calibrationCoefficients.t3)) >> 14

	tFine = var1 + var2
	T = (tFine*5 + 128) >> 8
	return
}

// calculatePressure uses the data slice and applies calibrations values on it to convert the value to an useful integer
func (d *Device) calculatePressure(data [8]byte, tFine int32) int32 {

	rawPressure := int32((((uint32(data[0]) << 8) | uint32(data[1])) << 8) | uint32(data[2]))
	rawPressure = rawPressure >> 4

	var1 := int64(tFine) - 128000
	var2 := var1 * var1 * int64(d.calibrationCoefficients.p6)
	var2 = var2 + ((var1 * int64(d.calibrationCoefficients.p5)) << 17)
	var2 = var2 + (int64(d.calibrationCoefficients.p4) << 35)
	var1 = ((var1 * var1 * int64(d.calibrationCoefficients.p3)) >> 8) + ((var1 * int64(d.calibrationCoefficients.p2)) << 12)
	var1 = ((int64(1) << 47) + var1) * int64(d.calibrationCoefficients.p1) >> 33

	if var1 == 0 {
		return 0 // avoid exception caused by division by zero
	}
	p := int64(1048576 - rawPressure)
	p = (((p << 31) - var2) * 3125) / var1
	var1 = (int64(d.calibrationCoefficients.p9) * (p >> 13) * (p >> 13)) >> 25
	var2 = (int64(d.calibrationCoefficients.p8) * p) >> 19

	p = ((p + var1 + var2) >> 8) + (int64(d.calibrationCoefficients.p7) << 4)
	p = (p / 256) * 1000
	return int32(p)
}

func (d *Device) calculateHumidity(data [8]byte, tFine int32) int32 {

	rawHumidity := int32(readInt(data[6], data[7]))

	v_x1_u32r := tFine - 76800

	v_x1_u32r = ((((rawHumidity << 14) - (int32(d.calibrationCoefficients.h4) << 20) -
		(int32(d.calibrationCoefficients.h5) * v_x1_u32r)) + 16384) >> 15) *
		(((((v_x1_u32r*int32(d.calibrationCoefficients.h6)>>10)*
			((v_x1_u32r*int32(d.calibrationCoefficients.h3)>>11)+32768)>>10)+
			2097152)*int32(d.calibrationCoefficients.h2) + 8192) >> 14)

	v_x1_u32r = (v_x1_u32r - (((((v_x1_u32r >> 15) * (v_x1_u32r >> 15)) >> 7) *
		int32(d.calibrationCoefficients.h1)) >> 4))

	println(v_x1_u32r)

	if v_x1_u32r < 0 {
		v_x1_u32r = 0
	}

	if v_x1_u32r > 419430400 {
		v_x1_u32r = 419430400
	}
	h := float32(v_x1_u32r >> 12)
	println(h)
	return int32(h)

	// var h float32

	// rawHumidity := int32(readInt(data[6], data[7]))
	// h = float32(tFine) - 76800

	// if h == 0 {
	// 	return 0 // TODO err is 'invalid data' from Bosch - include errors or not?
	// }

	// x := float32(rawHumidity) - (float32(d.calibrationCoefficients.h4)*64.0 +
	// 	(float32(d.calibrationCoefficients.h5) / 16384.0 * h))

	// y := float32(d.calibrationCoefficients.h2) / 65536.0 *
	// 	(1.0 + float32(d.calibrationCoefficients.h6)/67108864.0*h*
	// 		(1.0+float32(d.calibrationCoefficients.h3)/67108864.0*h))

	// h = x * y
	// h = h * (1 - float32(d.calibrationCoefficients.h1)*h/524288)
	// println(h)
	// return int32(h)
}
