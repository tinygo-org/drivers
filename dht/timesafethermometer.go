//go:build tinygo

// Package dht provides a driver for DHTXX family temperature and humidity sensors.
//
// [1] Datasheet DHT11: https://www.mouser.com/datasheet/2/758/DHT11-Technical-Data-Sheet-Translated-Version-1143054.pdf
// [2] Datasheet DHT22: https://cdn-shop.adafruit.com/datasheets/Digital+humidity+and+temperature+sensor+AM2302.pdf
// Adafruit C++ driver: https://github.com/adafruit/DHT-sensor-library

package dht // import "tinygo.org/x/drivers/dht"

import (
	"machine"
	"time"
)

// Device interface provides main functionality of the DHTXX sensors.
type Device interface {
	DummyDevice
	Configure(policy UpdatePolicy)
}

// managedDevice struct provides time control and optional automatic data retrieval from the sensor.
// It delegates all the functionality to device
type managedDevice struct {
	t          device
	lastUpdate time.Time
	policy     UpdatePolicy
}

// Measurements returns both measurements: temperature and humidity as they sent by the device.
// Depending on the UpdatePolicy of the device may update cached measurements.
func (m *managedDevice) Measurements() (temperature int16, humidity uint16, err error) {
	err = m.checkForUpdateOnDataRequest()
	if err != nil {
		return 0, 0, err
	}
	return m.t.Measurements()
}

// Getter for temperature. Temperature method returns temperature as it is sent by device.
// The temperature is measured temperature in Celsius multiplied by 10.
// Depending on the UpdatePolicy of the device may update cached measurements.
func (m *managedDevice) Temperature() (temp int16, err error) {
	err = m.checkForUpdateOnDataRequest()
	if err != nil {
		return 0, err
	}
	temp, err = m.t.Temperature()
	return
}

func (m *managedDevice) checkForUpdateOnDataRequest() (err error) {
	// update if necessary
	if m.policy.UpdateAutomatically {
		err = m.ReadMeasurements()
	}
	// ignore error if the data was updated recently
	// interface comparison does not work in tinygo. Therefore need to cast to explicit type
	if code, ok := err.(ErrorCode); ok && code == UpdateError {
		err = nil
	}
	// add error if the data is not initialized
	if !m.t.initialized {
		err = UninitializedDataError
	}
	return err
}

// Getter for temperature. TemperatureFloat returns temperature in a given scale.
// Depending on the UpdatePolicy of the device may update cached measurements.
func (m *managedDevice) TemperatureFloat(scale TemperatureScale) (float32, error) {
	err := m.checkForUpdateOnDataRequest()
	if err != nil {
		return 0, err
	}
	return m.t.TemperatureFloat(scale)
}

// Getter for humidity. Humidity returns humidity as it is sent by device.
// The humidity is measured in percentages multiplied by 10.
// Depending on the UpdatePolicy of the device may update cached measurements.
func (m *managedDevice) Humidity() (hum uint16, err error) {
	err = m.checkForUpdateOnDataRequest()
	if err != nil {
		return 0, err
	}
	return m.t.Humidity()
}

// Getter for humidity. HumidityFloat returns humidity in percentages.
// Depending on the UpdatePolicy of the device may update cached measurements.
func (m *managedDevice) HumidityFloat() (float32, error) {
	err := m.checkForUpdateOnDataRequest()
	if err != nil {
		return 0, err
	}
	return m.t.HumidityFloat()
}

// ReadMeasurements reads data from the sensor.
// The function will return UpdateError if it is called more frequently than specified in UpdatePolicy
func (m *managedDevice) ReadMeasurements() (err error) {
	timestamp := time.Now()
	if !m.t.initialized || timestamp.Sub(m.lastUpdate) > m.policy.UpdateTime {
		err = m.t.ReadMeasurements()
	} else {
		err = UpdateError
	}
	if err == nil {
		m.lastUpdate = timestamp
	}
	return
}

// Configure configures UpdatePolicy for Device.
// Configure checks for policy.UpdateTime and prevent from updating more frequently than specified in [1][2]
// to prevent undefined behaviour of the sensor.
func (m *managedDevice) Configure(policy UpdatePolicy) {
	if policy.UpdateAutomatically && policy.UpdateTime < time.Second*2 {
		policy.UpdateTime = time.Second * 2
	}
	m.policy = policy
}

// Constructor of the Device implementation.
// This implementation updates data every 2 seconds during data access.
func New(pin machine.Pin, deviceType DeviceType) Device {
	pin.High()
	return &managedDevice{
		t: device{
			pin:          pin,
			measurements: deviceType,
			initialized:  false,
		},
		lastUpdate: time.Time{},
		policy: UpdatePolicy{
			UpdateTime:          time.Second * 2,
			UpdateAutomatically: true,
		},
	}
}

// Constructor of the Device implementation with given UpdatePolicy
func NewWithPolicy(pin machine.Pin, deviceType DeviceType, updatePolicy UpdatePolicy) Device {
	pin.High()
	result := &managedDevice{
		t: device{
			pin:          pin,
			measurements: deviceType,
			initialized:  false,
		},
		lastUpdate: time.Time{},
	}
	result.Configure(updatePolicy)
	return result
}
