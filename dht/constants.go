package dht

import (
	"encoding/binary"
	"machine"
	"time"
)

type DeviceType uint8

func (d DeviceType) extractData(buf []byte) (temp int16, hum uint16) {
	if d == DHT11 {
		temp = int16(buf[2])
		if buf[3]&0x80 > 0 {
			temp = -1 - temp
		}
		temp *= 10
		temp += int16(buf[3] & 0x0f)
		hum = 10*uint16(buf[0]) + uint16(buf[1])
	} else {
		hum = binary.LittleEndian.Uint16(buf[0:2])
		temp = int16(buf[3])<<8 + int16(buf[2]&0x7f)
		if buf[2]&0x80 > 0 {
			temp = -temp
		}
	}
	return
}

type TemperatureScale uint8

func (t TemperatureScale) convertToFloat(temp int16) float32 {
	if t == C {
		return float32(temp) / 10
	} else {
		// Fahrenheit
		return float32(temp)*(9.0/50.) + 32.
	}
}

type ErrorCode uint8

const (
	startTimeout = time.Millisecond * 200
	startingLow  = time.Millisecond * 20

	DHT11 DeviceType = iota
	DHT22

	C TemperatureScale = iota
	F

	ChecksumError ErrorCode = iota
	NoSignalError
	NoDataError
	UpdateError
	UninitializedDataError
)

func (e ErrorCode) Error() string {
	switch e {
	case ChecksumError:
		return "checksum mismatch"
	case NoSignalError:
		return "no signal"
	case NoDataError:
		return "no data"
	case UpdateError:
		return "cannot update now"
	case UninitializedDataError:
		return "no measurements done"
	}
	// should never be reached
	return "unknown error"
}

// If update time is less than 2 seconds, thermometer will never update data automatically.
// It will require manual Update calls
type UpdatePolicy struct {
	UpdateTime          time.Duration
	UpdateAutomatically bool
}

var (
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
