//go:build tinygo

// Package dht provides a driver for DHTXX family temperature and humidity sensors.
//
// [1] Datasheet DHT11: https://www.mouser.com/datasheet/2/758/DHT11-Technical-Data-Sheet-Translated-Version-1143054.pdf
// [2] Datasheet DHT22: https://cdn-shop.adafruit.com/datasheets/Digital+humidity+and+temperature+sensor+AM2302.pdf
// Adafruit C++ driver: https://github.com/adafruit/DHT-sensor-library

package dht // import "tinygo.org/x/drivers/dht"

import (
	"machine"
	"runtime/interrupt"
	"time"
)

// DummyDevice provides a basic interface for DHT devices.
type DummyDevice interface {
	ReadMeasurements() error
	Measurements() (temperature int16, humidity uint16, err error)
	Temperature() (int16, error)
	TemperatureFloat(scale TemperatureScale) (float32, error)
	Humidity() (uint16, error)
	HumidityFloat() (float32, error)
}

// Basic implementation of the DummyDevice
// This implementation takes measurements from sensor only with ReadMeasurements function
// and does not provide a protection from too frequent calls for measurements.
// Since taking measurements from the sensor is time consuming procedure and blocks interrupts,
// user can avoid any hidden calls to the sensor.
type device struct {
	pin machine.Pin

	measurements DeviceType
	initialized  bool

	temperature int16
	humidity    uint16
}

// ReadMeasurements reads data from the sensor.
// According to documentation pin should be always, but the t *device restores pin to the state before call.
func (t *device) ReadMeasurements() error {
	// initial waiting
	state := powerUp(t.pin)
	defer t.pin.Set(state)
	err := t.read()
	if err == nil {
		t.initialized = true
	}
	return err
}

// Getter for temperature. Temperature method returns temperature as it is sent by device.
// The temperature is measured temperature in Celsius multiplied by 10.
// If no successful measurements for this device was performed, returns UninitializedDataError.
func (t *device) Temperature() (int16, error) {
	if !t.initialized {
		return 0, UninitializedDataError
	}
	return t.temperature, nil
}

// Getter for temperature. TemperatureFloat returns temperature in a given scale.
// If no successful measurements for this device was performed, returns UninitializedDataError.
func (t *device) TemperatureFloat(scale TemperatureScale) (float32, error) {
	if !t.initialized {
		return 0, UninitializedDataError
	}
	return scale.convertToFloat(t.temperature), nil
}

// Getter for humidity. Humidity returns humidity as it is sent by device.
// The humidity is measured in percentages multiplied by 10.
// If no successful measurements for this device was performed, returns UninitializedDataError.
func (t *device) Humidity() (uint16, error) {
	if !t.initialized {
		return 0, UninitializedDataError
	}
	return t.humidity, nil
}

// Getter for humidity. HumidityFloat returns humidity in percentages.
// If no successful measurements for this device was performed, returns UninitializedDataError.
func (t *device) HumidityFloat() (float32, error) {
	if !t.initialized {
		return 0, UninitializedDataError
	}
	return float32(t.humidity) / 10., nil
}

// Perform initialization of the communication protocol.
// Device lowers the voltage on pin for startingLow=20ms and starts listening for response
// Section 5.2 in [1]
func initiateCommunication(p machine.Pin) {
	// Send low signal to the device
	p.Configure(machine.PinConfig{Mode: machine.PinOutput})
	p.Low()
	time.Sleep(startingLow)
	// Set pin to high and wait for reply
	p.High()
	p.Configure(machine.PinConfig{Mode: machine.PinInput})
}

// Measurements returns both measurements: temperature and humidity as they sent by the device.
// If no successful measurements for this device was performed, returns UninitializedDataError.
func (t *device) Measurements() (temperature int16, humidity uint16, err error) {
	if !t.initialized {
		return 0, 0, UninitializedDataError
	}
	temperature = t.temperature
	humidity = t.humidity
	err = nil
	return
}

// Main routine that performs communication with the sensor
func (t *device) read() error {
	// initialize loop variables

	// buffer for the data sent by the sensor. Sensor sends 40 bits = 5 bytes
	bufferData := [5]byte{}
	buf := bufferData[:]

	// We perform measurements of the signal from the sensor by counting low and high cycles.
	// The bit is determined by the relative length of the high signal to low signal.
	// For 1, high signal will be longer than low, for 0---low is longer.
	// See section 5.3 [1]
	signalsData := [80]counter{}
	signals := signalsData[:]

	// Start communication protocol with sensor
	initiateCommunication(t.pin)
	// Wait for sensor's response and abort if sensor does not reply
	err := waitForDataTransmission(t.pin)
	if err != nil {
		return err
	}
	// count low and high cycles for sensor's reply
	receiveSignals(t.pin, signals)

	// process received signals and store the result in the buffer. Abort if data transmission was interrupted and not
	// all 40 bits were received
	err = t.extractData(signals[:], buf)
	if err != nil {
		return err
	}
	// Compute checksum and compare it to the one in data. Abort if checksum is incorrect
	if !isValid(buf[:]) {
		return ChecksumError
	}

	// Extract temperature and humidity data from buffer
	t.temperature, t.humidity = t.measurements.extractData(buf)
	return nil
}

// receiveSignals counts number of low and high cycles. The execution is time critical, so the function disables
// interrupts
func receiveSignals(pin machine.Pin, result []counter) {
	i := uint8(0)
	mask := interrupt.Disable()
	defer interrupt.Restore(mask)
	for ; i < 40; i++ {
		result[i*2] = expectChange(pin, false)
		result[i*2+1] = expectChange(pin, true)
	}
}

// extractData process signal counters and transforms them into bits.
// if any of the bits were not received (timed-out), returns NoDataError
func (t *device) extractData(signals []counter, buf []uint8) error {
	for i := uint8(0); i < 40; i++ {
		lowCycle := signals[i*2]
		highCycle := signals[i*2+1]
		if lowCycle == timeout || highCycle == timeout {
			return NoDataError
		}
		byteN := i >> 3
		buf[byteN] <<= 1
		if highCycle > lowCycle {
			buf[byteN] |= 1
		}
	}
	return nil
}

// waitForDataTransmission waits for reply from the sensor.
// If no reply received, returns NoSignalError.
// For more details, see section 5.2 in [1]
func waitForDataTransmission(p machine.Pin) error {
	// wait for thermometer to pull down
	if expectChange(p, true) == timeout {
		return NoSignalError
	}
	//wait for thermometer to pull up
	if expectChange(p, false) == timeout {
		return NoSignalError
	}
	// wait for thermometer to pull down and start sending the data
	if expectChange(p, true) == timeout {
		return NoSignalError
	}
	return nil
}

// Constructor function for a DummyDevice implementation.
// This device provides full control to the user.
// It does not do any hidden measurements calls and does not check
// for 2 seconds delay between measurements.
func NewDummyDevice(pin machine.Pin, deviceType DeviceType) DummyDevice {
	pin.High()
	return &device{
		pin:          pin,
		measurements: deviceType,
		initialized:  false,
		temperature:  0,
		humidity:     0,
	}
}
