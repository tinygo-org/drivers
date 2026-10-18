// Package bmi270 implements a driver for the Bosch BMI270 6-axis IMU,
// using either I2C or SPI.
//
// Datasheet: https://www.bosch-sensortec.com/media/boschsensortec/downloads/datasheets/bst-bmi270-ds000.pdf
package bmi270 // import "tinygo.org/x/drivers/bmi270"

import (
	_ "embed"
	"errors"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/pin"
)

// Config file from the Bosch BMI270 Sensor API, see LICENSE-config.
// https://github.com/boschsensortec/BMI270_SensorAPI/blob/master/bmi270.c
//
//go:embed bmi270-config.bin
var configFile string

// burstLen is the chunk size used to upload the config file.
const burstLen = 32

var (
	ErrDeviceNotConnected = errors.New("bmi270: device not connected")
	ErrInvalidConfig      = errors.New("bmi270: invalid config")
	ErrInitFailed         = errors.New("bmi270: initialization failed")
	ErrInvalidTemperature = errors.New("bmi270: invalid temperature")
)

// Device wraps an I2C or SPI connection to a BMI270 device.
type Device struct {
	bus     drivers.I2C
	Address uint16

	spi drivers.SPI
	cs  pin.Output

	accelRange uint8
	gyroRange  uint8
	data       [12]byte
	temp       int16
	buf        [burstLen + 1]byte
}

// NewI2C creates a new BMI270 connection over I2C. The bus must already be
// configured. It does not touch the device.
func NewI2C(bus drivers.I2C) Device {
	return Device{
		bus:        bus,
		Address:    Address,
		accelRange: uint8(AccelRange8G - 1),
		gyroRange:  uint8(GyroRange2000 - 1),
		temp:       temperatureInvalid,
	}
}

// NewSPI creates a new BMI270 connection over SPI. The bus and the cs pin
// must already be configured, with cs as an output. It does not touch the device.
func NewSPI(spi drivers.SPI, cs pin.Output) Device {
	return Device{
		spi:        spi,
		cs:         cs,
		accelRange: uint8(AccelRange8G - 1),
		gyroRange:  uint8(GyroRange2000 - 1),
		temp:       temperatureInvalid,
	}
}

// Configure resets the device, loads the config file, and enables the
// accelerometer, gyroscope and temperature sensor in performance mode.
func (d *Device) Configure(cfg Config) error {
	accelRange, accelConf, gyroRange, gyroConf, err := cfg.registers()
	if err != nil {
		return err
	}
	if !d.Connected() {
		return ErrDeviceNotConnected
	}

	// A soft reset is needed since INIT_CTRL may only be set once, see datasheet section 4.4.
	if err := d.writeRegister(RegCmd, cmdSoftReset); err != nil {
		return err
	}
	time.Sleep(2 * time.Millisecond)
	d.selectSPI()

	if err := d.writeRegister(RegPwrConf, 0); err != nil {
		return err
	}
	time.Sleep(450 * time.Microsecond)
	if err := d.writeRegister(RegInitCtrl, 0); err != nil {
		return err
	}
	if err := d.uploadConfig(); err != nil {
		return err
	}
	if err := d.writeRegister(RegInitCtrl, 1); err != nil {
		return err
	}
	if err := d.waitInit(); err != nil {
		return err
	}

	// Normal mode settings, see datasheet section 4.5.
	if err := d.writeRegister(RegPwrCtrl, pwrCtrlAccel|pwrCtrlGyro|pwrCtrlTemp); err != nil {
		return err
	}
	if err := d.writeRegister(RegAccelConf, accelConf); err != nil {
		return err
	}
	if err := d.writeRegister(RegAccelRange, accelRange); err != nil {
		return err
	}
	if err := d.writeRegister(RegGyroConf, gyroConf); err != nil {
		return err
	}
	if err := d.writeRegister(RegGyroRange, gyroRange); err != nil {
		return err
	}
	if err := d.writeRegister(RegPwrConf, pwrConfFIFOWakeup); err != nil {
		return err
	}

	d.accelRange = accelRange
	d.gyroRange = gyroRange
	return nil
}

// Connected reports whether a BMI270 answers.
func (d *Device) Connected() bool {
	d.selectSPI()
	data, err := d.readRegisters(RegChipID, 1)
	return err == nil && data[0] == ChipID
}

// Reset performs a soft reset. Call Configure to use the device again.
func (d *Device) Reset() error {
	if err := d.writeRegister(RegCmd, cmdSoftReset); err != nil {
		return err
	}
	d.accelRange = uint8(AccelRange8G - 1)
	d.gyroRange = uint8(GyroRange2000 - 1)
	return nil
}

// Update reads the requested measurements from the device.
func (d *Device) Update(which drivers.Measurement) error {
	if which&(drivers.Acceleration|drivers.AngularVelocity) != 0 {
		data, err := d.readRegisters(RegAccelData, len(d.data))
		if err != nil {
			return err
		}
		copy(d.data[:], data)
	}
	if which&drivers.Temperature != 0 {
		data, err := d.readRegisters(RegTemperature, 2)
		if err != nil {
			return err
		}
		d.temp = int16(uint16(data[0]) | uint16(data[1])<<8)
		if d.temp == temperatureInvalid {
			return ErrInvalidTemperature
		}
	}
	return nil
}

// Acceleration returns the last read acceleration in µg.
func (d *Device) Acceleration() (x, y, z int32) {
	return d.accel(0), d.accel(2), d.accel(4)
}

// AngularVelocity returns the last read angular velocity in µ°/s.
func (d *Device) AngularVelocity() (x, y, z int32) {
	return d.gyro(6), d.gyro(8), d.gyro(10)
}

// Temperature returns the last read temperature in milli °C.
func (d *Device) Temperature() int32 {
	// 0x0000 is 23 °C with 1/512 K per LSB, see datasheet section 5.2.34.
	return int32(d.temp)*125/64 + 23000
}

// ReadAcceleration reads the acceleration in µg.
func (d *Device) ReadAcceleration() (x, y, z int32, err error) {
	if err = d.Update(drivers.Acceleration); err != nil {
		return
	}
	x, y, z = d.Acceleration()
	return
}

// ReadRotation reads the angular velocity in µ°/s.
func (d *Device) ReadRotation() (x, y, z int32, err error) {
	if err = d.Update(drivers.AngularVelocity); err != nil {
		return
	}
	x, y, z = d.AngularVelocity()
	return
}

// ReadTemperature reads the temperature in milli °C.
// It is only valid while the gyroscope or accelerometer is enabled.
func (d *Device) ReadTemperature() (int32, error) {
	if err := d.Update(drivers.Temperature); err != nil {
		return 0, err
	}
	return d.Temperature(), nil
}

// DataReady reports whether new accelerometer and gyroscope data is available.
func (d *Device) DataReady() (accel, gyro bool, err error) {
	data, err := d.readRegisters(RegStatus, 1)
	if err != nil {
		return false, false, err
	}
	return data[0]&statusAccelReady != 0, data[0]&statusGyroReady != 0, nil
}

func (d *Device) accel(i int) int32 {
	// Full scale is 2g << accelRange over 32768 LSB.
	return int32(d.raw(i)) * 15625 / (256 >> d.accelRange)
}

func (d *Device) gyro(i int) int32 {
	// Full scale is 2000 dps >> gyroRange over 32768 LSB.
	return int32(int64(d.raw(i)) * int64(2000>>d.gyroRange) * 15625 / 512)
}

func (d *Device) raw(i int) int16 {
	return int16(uint16(d.data[i]) | uint16(d.data[i+1])<<8)
}

// selectSPI switches the interface to SPI after power up or soft reset,
// see datasheet section 4.4. The idle time is from section 6.4.
func (d *Device) selectSPI() {
	if d.spi == nil {
		return
	}
	d.readRegisters(RegChipID, 1)
	time.Sleep(450 * time.Microsecond)
}

func (d *Device) uploadConfig() error {
	for i := 0; i < len(configFile); i += burstLen {
		// INIT_ADDR counts in words, see datasheet section 4.4 footnote 4.
		d.buf[0] = RegInitAddr0
		d.buf[1] = byte(i/2) & 0x0F
		d.buf[2] = byte(i / 2 >> 4)
		if err := d.write(2); err != nil {
			return err
		}
		d.buf[0] = RegInitData
		n := copy(d.buf[1:], configFile[i:])
		if err := d.write(n); err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) waitInit() error {
	// Initialization takes at most 20 ms, see datasheet section 4.4.
	for i := 0; i < 30; i++ {
		time.Sleep(time.Millisecond)
		data, err := d.readRegisters(RegInternalStatus, 1)
		if err != nil {
			return err
		}
		if data[0]&internalStatusMask == internalStatusInitOK {
			return nil
		}
	}
	return ErrInitFailed
}

// readRegisters returns n bytes starting at reg. The result is only valid
// until the next bus operation.
func (d *Device) readRegisters(reg uint8, n int) ([]byte, error) {
	if d.spi != nil {
		// SPI reads start with a dummy byte, see datasheet section 3.
		b := d.buf[:n+2]
		b[0] = reg | spiRead
		for i := 1; i < len(b); i++ {
			b[i] = 0
		}
		d.cs.Set(false)
		err := d.spi.Tx(b, b)
		d.cs.Set(true)
		return b[2:], err
	}
	d.buf[0] = reg
	b := d.buf[1 : n+1]
	return b, d.bus.Tx(d.Address, d.buf[:1], b)
}

func (d *Device) writeRegister(reg, val uint8) error {
	d.buf[0] = reg
	d.buf[1] = val
	return d.write(1)
}

// write sends d.buf[0] as the register followed by n data bytes.
func (d *Device) write(n int) error {
	if d.spi != nil {
		d.buf[0] &^= spiRead
		d.cs.Set(false)
		err := d.spi.Tx(d.buf[:n+1], nil)
		d.cs.Set(true)
		return err
	}
	return d.bus.Tx(d.Address, d.buf[:n+1], nil)
}
