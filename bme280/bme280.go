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

	//d.Reset()
	//time.Sleep(5000 * time.Millisecond)

	data := make([]byte, 24)
	err := d.bus.ReadRegister(uint8(d.Address), REG_CALIBRATION, data)
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

// Temperature returns the temperature in celsius milli degrees (ºC/10)
func (d *Device) ReadTemperature() (int32, error) {
	rawTemp := d.rawTemp()

	temp, _ := d.calculateTemp(rawTemp)
	return temp, nil
}

// rawTemp returns the sensor's raw values of the temperature
func (d *Device) rawTemp() int32 {
	data, err := d.readData()
	if err != nil {
		return 0
	}

	return int32((((uint32(data[3]) << 8) | uint32(data[4])) << 8) | uint32(data[5]))
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
func (d *Device) readData() ([]byte, error) {
	data := make([]byte, 8)
	err := d.bus.ReadRegister(uint8(d.Address), REG_PRESSURE, data)
	if err != nil {
		println(err)
		return nil, err
	}
	return data, nil
}

func (d *Device) calculateTemp(rawTemp int32) (int32, int32) {
	temp := rawTemp >> 4

	var1 := (((temp >> 3) - (int32(d.calibrationCoefficients.t1) << 1)) * int32(d.calibrationCoefficients.t2)) >> 11
	var2 := (((((temp >> 4) - int32(d.calibrationCoefficients.t1)) * ((temp >> 4) - int32(d.calibrationCoefficients.t1))) >> 12) * int32(d.calibrationCoefficients.t3)) >> 14

	tFine := var1 + var2
	T := (tFine*5 + 128) >> 8

	return T, tFine
}
