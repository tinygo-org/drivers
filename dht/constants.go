package dht

import (
	"encoding/binary"
	"errors"
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

const (
	startTimeout = time.Millisecond * 200
	startingLow  = time.Millisecond * 20

	DHT11 DeviceType = iota
	DHT22

	C TemperatureScale = iota
	F
)

var (
	timeout uint16

	checksumError = errors.New("checksum mismatch")
	noSignalError = errors.New("no signal")
	noDataError   = errors.New("no data")
)

func init() {
	timeout = cyclesPerMillisecond()
}

func cyclesPerMillisecond() uint16 {
	freq := machine.CPUFrequency()
	freq /= 1000
	return uint16(freq)
}
