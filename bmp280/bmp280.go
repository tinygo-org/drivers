package bmp280

import (
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/legacy"
)

// OversamplingMode is the oversampling ratio of the temperature or pressure measurement.
type Oversampling uint

// Mode is the Power Mode.
type Mode uint

// Standby is the inactive period between the reads when the sensor is in normal power mode.
type Standby uint

// Filter unwanted changes in measurement caused by external (environmental) or internal changes (IC).
type Filter uint

// Device wraps an I2C connection to a BMP280 device.
type Device struct {
	bus         drivers.I2C
	Address     uint16
	cali        calibrationCoefficients
	Temperature Oversampling
	Pressure    Oversampling
	Mode        Mode
	Standby     Standby
	Filter      Filter
}

type calibrationCoefficients struct {
	// Temperature compensation
	t1 uint16
	t2 int16
	t3 int16

	// Pressure compensation
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

// New creates a new BMP280 connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not initialize the device.
// You must call Configure() first in order to use the device itself.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
	}
}

// Connected returns whether a BMP280 has been found.
// It does a "who am I" request and checks the response.
func (d *Device) Connected() bool {
	data := make([]byte, 1)
	legacy.ReadRegister(d.bus, uint8(d.Address), REG_ID, data)
	return data[0] == CHIP_ID
}

// Reset preforms complete power-on-reset procedure.
// It is required to call Configure afterwards.
func (d *Device) Reset() {
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_RESET, []byte{CMD_RESET})
}

// Configure sets up the device for communication and
// read the calibration coefficients.
func (d *Device) Configure(standby Standby, filter Filter, temp Oversampling, pres Oversampling, mode Mode) {
	d.Standby = standby
	d.Filter = filter
	d.Temperature = temp
	d.Pressure = pres
	d.Mode = mode

	//  Write the configuration (standby, filter, spi 3 wire)
	config := uint(d.Standby<<5) | uint(d.Filter<<2) | 0x00
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_CONFIG, []byte{byte(config)})

	// Write the control (temperature oversampling, pressure oversampling,
	config = uint(d.Temperature<<5) | uint(d.Pressure<<2) | uint(d.Mode)
	legacy.WriteRegister(d.bus, uint8(d.Address), REG_CTRL_MEAS, []byte{byte(config)})

	// Read Calibration data
	data := make([]byte, 24)
	err := legacy.ReadRegister(d.bus, uint8(d.Address), REG_CALI, data)
	if err != nil {
		return
	}

	// Datasheet: 3.11.2 Trimming parameter readout
	d.cali.t1 = readUintLE(data[0], data[1])
	d.cali.t2 = readIntLE(data[2], data[3])
	d.cali.t3 = readIntLE(data[4], data[5])

	d.cali.p1 = readUintLE(data[6], data[7])
	d.cali.p2 = readIntLE(data[8], data[9])
	d.cali.p3 = readIntLE(data[10], data[11])
	d.cali.p4 = readIntLE(data[12], data[13])
	d.cali.p5 = readIntLE(data[14], data[15])
	d.cali.p6 = readIntLE(data[16], data[17])
	d.cali.p7 = readIntLE(data[18], data[19])
	d.cali.p8 = readIntLE(data[20], data[21])
	d.cali.p9 = readIntLE(data[22], data[23])
}

// PrintCali prints the Calibration information.
func (d *Device) PrintCali() {
	println("T1:", d.cali.t1)
	println("T2:", d.cali.t2)
	println("T3:", d.cali.t3)

	println("P1:", d.cali.p1)
	println("P2:", d.cali.p2)
	println("P3:", d.cali.p3)
	println("P4:", d.cali.p4)
	println("P5:", d.cali.p5)
	println("P6:", d.cali.p6)
	println("P7:", d.cali.p7)
	println("P8:", d.cali.p8)
	println("P9:", d.cali.p9, "\n")
}

// ReadTemperature returns the temperature in celsius milli degrees (°C/1000).
func (d *Device) ReadTemperature() (temperature int32, err error) {
	data, err := d.readData(REG_TEMP, 3)
	if err != nil {
		return
	}

	rawTemp := convert3Bytes(data[0], data[1], data[2])

	// Datasheet: 8.2 Compensation formula in 32 bit fixed point
	// Temperature compensation
	var1 := ((rawTemp >> 3) - int32(d.cali.t1<<1)) * int32(d.cali.t2) >> 11
	var2 := (((rawTemp >> 4) - int32(d.cali.t1)) * ((rawTemp >> 4) - int32(d.cali.t1)) >> 12) *
		int32(d.cali.t3) >> 14

	tFine := var1 + var2

	// Convert from degrees to milli degrees by multiplying by 10.
	// Will output 30250 milli degrees celsius for 30.25 degrees celsius
	temperature = 10 * ((tFine*5 + 128) >> 8)
	return
}

// ReadPressure returns the pressure in milli pascals (mPa).
func (d *Device) ReadPressure() (pressure int32, err error) {
	// First 3 bytes are Pressure, last 3 bytes are Temperature
	data, err := d.readData(REG_PRES, 6)
	if err != nil {
		return
	}

	rawTemp := convert3Bytes(data[3], data[4], data[5])

	// Datasheet: 8.2 Compensation formula in 32 bit fixed point
	// Calculate tFine (temperature), used for the Pressure compensation
	var1 := ((rawTemp >> 3) - int32(d.cali.t1<<1)) * int32(d.cali.t2) >> 11
	var2 := (((rawTemp >> 4) - int32(d.cali.t1)) * ((rawTemp >> 4) - int32(d.cali.t1)) >> 12) *
		int32(d.cali.t3) >> 14

	tFine := var1 + var2

	rawPres := convert3Bytes(data[0], data[1], data[2])

	// Datasheet: 8.2 Compensation formula in 32 bit fixed point
	// Pressure compensation
	var1 = (tFine >> 1) - 64000
	var2 = (((var1 >> 2) * (var1 >> 2)) >> 11) * int32(d.cali.p6)
	var2 = var2 + ((var1 * int32(d.cali.p5)) << 1)
	var2 = (var2 >> 2) + (int32(d.cali.p4) << 16)
	var1 = (((int32(d.cali.p3) * (((var1 >> 2) * (var1 >> 2)) >> 13)) >> 3) +
		((int32(d.cali.p2) * var1) >> 1)) >> 18
	var1 = ((32768 + var1) * int32(d.cali.p1)) >> 15

	if var1 == 0 {
		return 0, nil
	}

	p := uint32(((1048576 - rawPres) - (var2 >> 12)) * 3125)
	if p < 0x80000000 {
		p = (p << 1) / uint32(var1)
	} else {
		p = (p / uint32(var1)) * 2
	}

	var1 = (int32(d.cali.p9) * int32(((p>>3)*(p>>3))>>13)) >> 12
	var2 = (int32(p>>2) * int32(d.cali.p8)) >> 13

	return 1000 * (int32(p) + ((var1 + var2 + int32(d.cali.p7)) >> 4)), nil
}

// readData reads n number of bytes of the specified register
func (d *Device) readData(register int, n int) ([]byte, error) {
	// If not in normal mode, set the mode to FORCED mode, to prevent incorrect measurements
	// After the measurement in FORCED mode, the sensor will return to SLEEP mode
	if d.Mode != MODE_NORMAL {
		config := uint(d.Temperature<<5) | uint(d.Pressure<<2) | uint(MODE_FORCED)
		legacy.WriteRegister(d.bus, uint8(d.Address), REG_CTRL_MEAS, []byte{byte(config)})
	}

	// Check STATUS register, wait if data is not available yet
	status := make([]byte, 1)
	for legacy.ReadRegister(d.bus, uint8(d.Address), uint8(REG_STATUS), status[0:]); status[0] != 4 && status[0] != 0; legacy.ReadRegister(d.bus, uint8(d.Address), uint8(REG_STATUS), status[0:]) {
		time.Sleep(time.Millisecond)
	}

	// Read the requested register
	data := make([]byte, n)
	err := legacy.ReadRegister(d.bus, uint8(d.Address), uint8(register), data[:])
	return data, err
}

// convert3Bytes converts three bytes to int32
func convert3Bytes(msb byte, b1 byte, lsb byte) int32 {
	return int32(((((uint32(msb) << 8) | uint32(b1)) << 8) | uint32(lsb)) >> 4)
}

// readUint converts two bytes to uint16
func readUint(msb byte, lsb byte) uint16 {
	return (uint16(msb) << 8) | uint16(lsb)
}

// readUintLE converts two little endian bytes to uint16
func readUintLE(msb byte, lsb byte) uint16 {
	temp := readUint(msb, lsb)
	return (temp >> 8) | (temp << 8)
}

// readIntLE converts two little endian bytes to int16
func readIntLE(msb byte, lsb byte) int16 {
	return int16(readUintLE(msb, lsb))
}
