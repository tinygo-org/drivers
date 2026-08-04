// Package scd30 provides a driver for the Sensirion SCD30 CO2, temperature,
// and humidity sensor.
//
// Datasheet: https://sensirion.com/media/documents/D7CEEF4A/6165372F/Sensirion_CO2_Sensors_SCD30_Interface_Description.pdf
package scd30 // import "tinygo.org/x/drivers/scd30"

import (
	"encoding/binary"
	"errors"
	"math"
	"time"

	"tinygo.org/x/drivers"
)

const readDelay = 4 * time.Millisecond

var (
	ErrCRC = errors.New("scd30: invalid CRC")

	ErrInvalidInterval = errors.New("scd30: measurement interval must be between 2 and 1800 seconds")

	ErrInvalidAmbientPressure = errors.New("scd30: ambient pressure must be zero or between 700 and 1400 mbar")
)

// Config contains the SCD30 continuous measurement configuration.
type Config struct {
	// MeasurementInterval is the interval between measurements in seconds and
	// must be between 2 and 1800.
	MeasurementInterval uint16

	// AutomaticSelfCalibration enables or disables automatic self-calibration.
	AutomaticSelfCalibration bool
}

// DefaultConfig contains the power-on defaults documented for the SCD30.
var DefaultConfig = Config{
	MeasurementInterval:      2,
	AutomaticSelfCalibration: false,
}

// Device is a Sensirion SCD30 sensor connected over I2C.
type Device struct {
	bus drivers.I2C
	tx  [5]byte
	rx  [18]byte

	co2         int32
	temperature int32
	humidity    int32
}

var _ drivers.Sensor = (*Device)(nil)

// New returns a new SCD30 driver. It performs no I/O.
func New(bus drivers.I2C) *Device {
	return &Device{bus: bus}
}

// Configure applies the continuous measurement interval and automatic
// self-calibration settings. It does not start continuous measurement.
func (d *Device) Configure(config Config) error {
	if err := d.SetMeasurementInterval(config.MeasurementInterval); err != nil {
		return err
	}
	return d.SetAutomaticSelfCalibration(config.AutomaticSelfCalibration)
}

// Connected reports whether an SCD30 responds with a valid data-ready status.
func (d *Device) Connected() bool {
	_, err := d.DataReady()
	return err == nil
}

// SetMeasurementInterval sets the continuous measurement interval in seconds.
func (d *Device) SetMeasurementInterval(seconds uint16) error {
	if seconds < minimumMeasurementInterval || seconds > maximumMeasurementInterval {
		return ErrInvalidInterval
	}
	return d.writeCommandWithArgument(commandSetMeasurementInterval, seconds)
}

// SetAutomaticSelfCalibration enables or disables automatic self-calibration.
func (d *Device) SetAutomaticSelfCalibration(enabled bool) error {
	var value uint16
	if enabled {
		value = 1
	}
	return d.writeCommandWithArgument(commandSetAutoCalibration, value)
}

// StartContinuousMeasurement begins periodic measurements. Ambient pressure
// must be zero to disable pressure compensation, or between 700 and 1400 mbar.
func (d *Device) StartContinuousMeasurement(ambientPressure uint16) error {
	if ambientPressure != 0 && (ambientPressure < minimumAmbientPressure || ambientPressure > maximumAmbientPressure) {
		return ErrInvalidAmbientPressure
	}
	return d.writeCommandWithArgument(commandStartContinuousMeasurement, ambientPressure)
}

// StopContinuousMeasurement stops periodic measurements.
func (d *Device) StopContinuousMeasurement() error {
	return d.writeCommand(commandStopContinuousMeasurement)
}

// DataReady reports whether a new measurement can be read.
func (d *Device) DataReady() (bool, error) {
	if err := d.readCommand(commandDataReady, d.rx[:3]); err != nil {
		return false, err
	}
	value, err := decodeWord(d.rx[:3])
	if err != nil {
		return false, err
	}
	return value != 0, nil
}

// ReadMeasurement reads and caches the latest CO2, temperature, and humidity
// measurement. Use DataReady before calling ReadMeasurement.
func (d *Device) ReadMeasurement() error {
	if err := d.readCommand(commandReadMeasurement, d.rx[:18]); err != nil {
		return err
	}

	var data [12]byte
	for source, destination := 0, 0; source < 18; source, destination = source+3, destination+2 {
		value, err := decodeWord(d.rx[source : source+3])
		if err != nil {
			return err
		}
		binary.BigEndian.PutUint16(data[destination:destination+2], value)
	}

	co2 := decodeFloat32(data[0:4])
	temperature := decodeFloat32(data[4:8])
	humidity := decodeFloat32(data[8:12])

	d.co2 = roundFixed(co2, 1)
	d.temperature = roundFixed(temperature, 1000)
	d.humidity = roundFixed(humidity, 100)
	return nil
}

// Update reads and caches all measurements if any supported measurement was
// requested. The SCD30 provides all three values in a single transaction.
func (d *Device) Update(which drivers.Measurement) error {
	if which&(drivers.Concentration|drivers.Temperature|drivers.Humidity) == 0 {
		return nil
	}
	return d.ReadMeasurement()
}

// CO2 returns the last read CO2 concentration in parts per million.
func (d *Device) CO2() int32 {
	return d.co2
}

// Temperature returns the last read temperature in millidegrees Celsius.
func (d *Device) Temperature() int32 {
	return d.temperature
}

// Humidity returns the last read relative humidity in hundredths of a percent.
func (d *Device) Humidity() int32 {
	return d.humidity
}

func (d *Device) readCommand(command uint16, response []byte) error {
	if err := d.writeCommand(command); err != nil {
		return err
	}
	// The datasheet requires a delay greater than 3ms before reading.
	time.Sleep(readDelay)
	return d.bus.Tx(Address, nil, response)
}

func (d *Device) writeCommand(command uint16) error {
	binary.BigEndian.PutUint16(d.tx[:2], command)
	return d.bus.Tx(Address, d.tx[:2], nil)
}

func (d *Device) writeCommandWithArgument(command, argument uint16) error {
	binary.BigEndian.PutUint16(d.tx[:2], command)
	binary.BigEndian.PutUint16(d.tx[2:4], argument)
	d.tx[4] = crc8(d.tx[2:4])
	return d.bus.Tx(Address, d.tx[:5], nil)
}

func decodeWord(data []byte) (uint16, error) {
	if len(data) != 3 || crc8(data[:2]) != data[2] {
		return 0, ErrCRC
	}
	return binary.BigEndian.Uint16(data[:2]), nil
}

func decodeFloat32(data []byte) float32 {
	return math.Float32frombits(binary.BigEndian.Uint32(data))
}

func roundFixed(value float32, scale int32) int32 {
	scaled := value * float32(scale)
	if scaled < 0 {
		return int32(scaled - 0.5)
	}
	return int32(scaled + 0.5)
}

func crc8(data []byte) byte {
	value := byte(0xff)
	for _, current := range data {
		value ^= current
		for bit := 0; bit < 8; bit++ {
			if value&0x80 != 0 {
				value = value<<1 ^ 0x31
			} else {
				value <<= 1
			}
		}
	}
	return value
}
