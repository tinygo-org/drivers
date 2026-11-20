// Package bmp580 provides a driver for the Bosch BMP580 barometric
// pressure sensor over I2C.
//
// Datasheet: https://cdn-shop.adafruit.com/product-files/6411/BMP580.pdf
package bmp580 // import "tinygo.org/x/drivers/bmp580"

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

var (
	ErrNotReady = errors.New("bmp580: sensor not ready after reset")
	ErrTimeout  = errors.New("bmp580: measurement did not finish")
)

// Polling interval and limit while a forced measurement runs. With 128
// samples of each a measurement takes about 0.1 s, datasheet Table 4.
const (
	pollInterval = 5 * time.Millisecond
	maxPolls     = 100
)

// Device wraps an I2C connection to a BMP580 device.
type Device struct {
	bus         drivers.I2C
	Address     uint16
	temperature int32
	pressure    int32
	buf         [7]byte
}

// Config sets the oversampling of each measurement. The zero value takes a
// single sample of each.
type Config struct {
	Pressure    Oversampling
	Temperature Oversampling
}

// New creates a new BMP580 connection. The I2C bus must already be configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C) Device {
	return Device{bus: bus, Address: Address}
}

// Connected returns whether a BMP580 has been found.
func (d *Device) Connected() bool {
	id, err := d.readReg(regChipID)
	return err == nil && id == chipIDBMP580
}

// Configure resets the sensor, checks that it is ready and sets the
// oversampling. The sensor is left in standby until the next Update.
func (d *Device) Configure(cfg Config) error {
	// The soft reset is not ACKed over I2C, datasheet section 7.36. The
	// INT_STATUS check below confirms it worked.
	d.writeReg(regCmd, cmdSoftReset)
	time.Sleep(2 * time.Millisecond)
	// Reading INT_STATUS also clears it.
	intStatus, err := d.readReg(regIntStatus)
	if err != nil {
		return err
	}
	status, err := d.readReg(regStatus)
	if err != nil {
		return err
	}
	if intStatus&intStatusPOR == 0 || status&statusNVMReady == 0 || status&statusNVMError != 0 {
		return ErrNotReady
	}
	if err := d.writeReg(regODRConfig, odrDeepDisable|odrReset1Hz|pwrModeStandby); err != nil {
		return err
	}
	osr := osrPressureEnable | byte(cfg.Pressure&7)<<3 | byte(cfg.Temperature&7)
	return d.writeReg(regOSRConfig, osr)
}

// Update takes one measurement. The sensor measures temperature and pressure
// together, so which only has to include drivers.Temperature or
// drivers.Pressure.
func (d *Device) Update(which drivers.Measurement) error {
	if which&(drivers.Temperature|drivers.Pressure) == 0 {
		return nil
	}
	if err := d.writeReg(regODRConfig, odrDeepDisable|odrReset1Hz|pwrModeForced); err != nil {
		return err
	}
	// The sensor goes back to standby when the measurement is done.
	for polls := 0; ; polls++ {
		time.Sleep(pollInterval)
		mode, err := d.readReg(regODRConfig)
		if err != nil {
			return err
		}
		if mode&pwrModeMask == pwrModeStandby {
			break
		}
		if polls == maxPolls {
			return ErrTimeout
		}
	}
	d.buf[0] = regTempData
	if err := d.bus.Tx(d.Address, d.buf[:1], d.buf[1:7]); err != nil {
		return err
	}
	d.temperature, d.pressure = convert(d.buf[1:7])
	return nil
}

// Temperature returns the temperature of the last Update in milli degrees
// Celsius: 21500 is 21.5 °C.
func (d *Device) Temperature() int32 {
	return d.temperature
}

// Pressure returns the pressure of the last Update in milli pascal:
// 101325000 is 1013.25 hPa.
func (d *Device) Pressure() int32 {
	return d.pressure
}

// ReadTemperature takes a measurement and returns the temperature in milli
// degrees Celsius.
func (d *Device) ReadTemperature() (int32, error) {
	err := d.Update(drivers.Temperature)
	return d.temperature, err
}

// ReadPressure takes a measurement and returns the pressure in milli pascal.
func (d *Device) ReadPressure() (int32, error) {
	err := d.Update(drivers.Pressure)
	return d.pressure, err
}

// convert turns the six data registers into physical units. Temperature is a
// signed 24-bit value in 1/65536 °C, pressure an unsigned one in 1/64 Pa.
func convert(data []byte) (milliCelsius, milliPascal int32) {
	rawT := int32(uint32(data[0])<<8|uint32(data[1])<<16|uint32(data[2])<<24) >> 8
	rawP := uint32(data[3]) | uint32(data[4])<<8 | uint32(data[5])<<16
	// 2^24 * 125 still fits in an int32.
	return int32(int64(rawT) * 1000 >> 16), int32(rawP * 125 / 8)
}

func (d *Device) readReg(reg uint8) (byte, error) {
	d.buf[0] = reg
	err := d.bus.Tx(d.Address, d.buf[:1], d.buf[1:2])
	return d.buf[1], err
}

func (d *Device) writeReg(reg, value uint8) error {
	d.buf[0] = reg
	d.buf[1] = value
	return d.bus.Tx(d.Address, d.buf[:2], nil)
}
