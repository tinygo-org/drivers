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

// Celsius and Fahrenheit temperature scales
type TemperatureScale uint8

func (t TemperatureScale) convertToFloat(temp int16) float32 {
	if t == C {
		return float32(temp) / 10
	} else {
		// Fahrenheit
		return float32(temp)*(9.0/50.) + 32.
	}
}

// All functions return ErrorCode instance as error. This class can be used for more efficient error processing
type ErrorCode uint8

const (
	startTimeout = time.Millisecond * 200
	startingLow  = time.Millisecond * 20

	C TemperatureScale = iota
	F

	ChecksumError ErrorCode = iota
	NoSignalError
	NoDataError
	UpdateError
	UninitializedDataError
)

// error interface implementation for ErrorCode
func (e ErrorCode) Error() string {
	switch e {
	case ChecksumError:
		// DHT returns ChecksumError if all the data from the sensor was received, but the checksum does not match.
		return "checksum mismatch"
	case NoSignalError:
		// DHT returns NoSignalError if there was no reply from the sensor. Check sensor connection or the correct pin
		// sis chosen,
		return "no signal"
	case NoDataError:
		// DHT returns NoDataError if the connection was successfully initialized, but not all 40 bits from
		// the sensor is received
		return "no data"
	case UpdateError:
		// DHT returns UpdateError if ReadMeasurements function is called before time specified in UpdatePolicy or
		// less than 2 seconds after past measurement
		return "cannot update now"
	case UninitializedDataError:
		// DHT returns UninitializedDataError if user attempts to access data before first measurement
		return "no measurements done"
	}
	// should never be reached
	return "unknown error"
}

// Update policy of the DHT device. UpdateTime cannot be shorter than 2 seconds. According to dht specification sensor
// will return undefined data if update requested less than 2 seconds before last usage
type UpdatePolicy struct {
	UpdateTime          time.Duration
	UpdateAutomatically bool
}

var (
	// timeout counter equal to number of ticks per 1 millisecond
	timeout counter
)

func init() {
	timeout = cyclesPerMillisecond()
}

func cyclesPerMillisecond() counter {
	freq := machine.CPUFrequency()
	freq /= 1000
	return counter(freq)
}
