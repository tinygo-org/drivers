package bmi160

import (
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

// DeviceSPI is the SPI interface to a BMI160 accelerometer/gyroscope. There is
// also an I2C interface, but it is not yet supported.
type DeviceSPI struct {
	// Chip select pin
	CSB machine.Pin

	// SPI bus (requires chip select to be usable).
	Bus drivers.SPI
}

// NewSPI returns a new device driver. The pin and SPI interface are not
// touched, provide a fully configured SPI object and call Configure to start
// using this device.
func NewSPI(csb machine.Pin, spi drivers.SPI) *DeviceSPI {
	return &DeviceSPI{
		CSB: csb, // chip select
		Bus: spi,
	}
}

// Configure configures the BMI160 for use. It configures the CSB pin and
// configures the BMI160, but it does not configure the SPI interface (it is
// assumed to be up and running).
func (d *DeviceSPI) Configure() error {
	d.CSB.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.CSB.High()

	// The datasheet recommends doing a register read from address 0x7F to get
	// SPI communication going:
	// > If CSB sees a rising edge after power-up, the BMI160 interface switches
	// > to SPI until a reset or the next power-up occurs. Therefore, a CSB
	// > rising edge is needed before starting the SPI communication. Hence, it
	// > is recommended to perform a SPI single read access to the ADDRESS 0x7F
	// > before the actual communication in order to use the SPI interface.
	d.readRegister(0x7F)

	// Power up the accelerometer. 0b0001_00nn is the command format, with 0b01
	// indicating normal mode.
	d.runCommand(0b0001_0001)

	// Power up the gyroscope. 0b0001_01nn is the command format, with 0b01
	// indicating normal mode.
	d.runCommand(0b0001_0101)

	// Wait until the device is fully initialized. Even after the command has
	// finished, the gyroscope may not be fully powered on. Therefore, wait
	// until we get an expected value.
	// This takes 30ms or so.
	for {
		// Wait for the acc_pmu_status and gyr_pmu_status to both be 0b01.
		if d.readRegister(reg_PMU_STATUS) == 0b0001_0100 {
			break
		}
	}

	return nil
}

// Connected check whether the device appears to be properly connected. It reads
// the CHIPID, which must be 0xD1 for the BMI160.
func (d *DeviceSPI) Connected() bool {
	return d.readRegister(reg_CHIPID) == 0xD1
}

// Reset restores the device to the state after power up. This can be useful to
// easily disable the accelerometer and gyroscope to reduce current consumption.
func (d *DeviceSPI) Reset() error {
	d.runCommand(0xB6) // softreset
	return nil
}

// ReadTemperature returns the temperature in celsius milli degrees (°C/1000).
func (d *DeviceSPI) ReadTemperature() (temperature int32, err error) {
	data := []byte{0x80 | reg_TEMPERATURE_0, 0, 0}
	d.CSB.Low()
	err = d.Bus.Tx(data, data)
	d.CSB.High()
	if err != nil {
		return
	}
	rawTemperature := int16(uint16(data[1]) | uint16(data[2])<<8)
	// 0x0000 is 23°C
	// 0x7fff is ~87°C
	// We use 0x8000 instead of 0x7fff to make the formula easier. The result
	// should be near identical and shouldn't affect the result too much (the
	// temperature sensor has an offset of around 2°C so isn't very reliable).
	// So the formula is as follows:
	// 1. Scale from 0x0000..0x8000 to 0..(87-23).
	//    rawTemperature * (87-23) / 0x8000
	// 2. Convert to centidegrees.
	//    rawTemperature * 1000 * (87-23) / 0x8000
	// 3. Add 23°C offset.
	//    rawTemperature * 1000 * (87-23) / 0x8000 + 23000
	// 4. Simplify.
	//    rawTemperature * 1000 * 64 / 0x8000 + 23000
	//    rawTemperature * 64000 / 0x8000 + 23000
	//    rawTemperature * 125 / 64 + 23000
	temperature = int32(rawTemperature)*125/64 + 23000
	return
}

// ReadAcceleration reads the current acceleration from the device and returns
// it in µg (micro-gravity). When one of the axes is pointing straight to Earth
// and the sensor is not moving the returned value will be around 1000000 or
// -1000000.
func (d *DeviceSPI) ReadAcceleration() (x int32, y int32, z int32, err error) {
	data := []byte{0x80 | reg_ACC_XL, 0, 0, 0, 0, 0, 0}
	d.CSB.Low()
	err = d.Bus.Tx(data, data)
	d.CSB.High()
	if err != nil {
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
	x = int32(int16(uint16(data[1])|uint16(data[2])<<8)) * 15625 / 256
	y = int32(int16(uint16(data[3])|uint16(data[4])<<8)) * 15625 / 256
	z = int32(int16(uint16(data[5])|uint16(data[6])<<8)) * 15625 / 256
	return
}

// ReadRotation reads the current rotation from the device and returns it in
// µ°/s (micro-degrees/sec). This means that if you were to do a complete
// rotation along one axis and while doing so integrate all values over time,
// you would get a value close to 360000000.
func (d *DeviceSPI) ReadRotation() (x int32, y int32, z int32, err error) {
	data := []byte{0x80 | reg_GYR_XL, 0, 0, 0, 0, 0, 0}
	d.CSB.Low()
	err = d.Bus.Tx(data, data)
	d.CSB.High()
	if err != nil {
		return
	}
	// First the value is converted from a pair of bytes to a signed 16-bit
	// value and then to a signed 32-bit value to avoid integer overflow.
	// Then the value is scaled to µ°/s (micro-degrees per second).
	// The default is 2000°/s full scale range for -32768..32767.
	// The formula works as follows (taking X as an example):
	// 1. Scale from 32768 to 2000. This means that it is in °/s units.
	//    rawX * 2000 / 32768
	// 2. Scale to µ°/s by multiplying by 1e6.
	//    rawX * 1e6 * 2000 / 32768
	// 3. Simplify.
	//    rawX * 2e9 / 32768
	//    rawX * 1953125 / 32
	rawX := int32(int16(uint16(data[1]) | uint16(data[2])<<8))
	rawY := int32(int16(uint16(data[3]) | uint16(data[4])<<8))
	rawZ := int32(int16(uint16(data[5]) | uint16(data[6])<<8))
	x = int32(int64(rawX) * 1953125 / 32)
	y = int32(int64(rawY) * 1953125 / 32)
	z = int32(int64(rawZ) * 1953125 / 32)
	return
}

// runCommand runs a BMI160 command through the CMD register. It waits for the
// command to complete before returning.
func (d *DeviceSPI) runCommand(command uint8) {
	d.writeRegister(reg_CMD, command)
	for {
		response := d.readRegister(reg_CMD)
		if response == 0 {
			return // command was completed
		}
	}
}

// readRegister reads from a single BMI160 register. It should only be used for
// single register reads, not for reading multiple registers at once.
func (d *DeviceSPI) readRegister(address uint8) uint8 {
	// I don't know why but it appears necessary to sleep for a bit here.
	time.Sleep(time.Millisecond)

	data := []byte{0x80 | address, 0}
	d.CSB.Low()
	d.Bus.Tx(data, data)
	d.CSB.High()
	return data[1]
}

// writeRegister writes a single byte BMI160 register. It should only be used
// for writing to a single register.
func (d *DeviceSPI) writeRegister(address, data uint8) {
	// I don't know why but it appears necessary to sleep for a bit here.
	time.Sleep(time.Millisecond)

	d.CSB.Low()
	d.Bus.Tx([]byte{address, data}, []byte{0, 0})
	d.CSB.High()
}
