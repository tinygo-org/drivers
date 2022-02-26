package dht

import (
	"encoding/binary"
)

// DeviceType is the enum type for device type
type DeviceType uint8

const (
	DHT11 DeviceType = iota
	DHT22
)

// extractData parses information received from the sensor.
// The 2 first buffers are for the humidity and
// the 2 following corresponds to the temperature.
func (d DeviceType) extractData(buf []byte) (temp int16, hum uint16) {
	switch d {
	case DHT11:
		hum = 10*uint16(buf[0]) + uint16(buf[1])
		temp = int16(buf[2])
		if buf[3]&0x80 > 0 {
			temp = -1 - temp
		}
		temp *= 10
		temp += int16(buf[3] & 0x0f)
	case DHT22:
		hum = binary.BigEndian.Uint16(buf[0:2])
		temp = int16(buf[2]&0x7f)<<8 + int16(buf[3])
		// the first bit corresponds to the sign bit
		if buf[2]&0x80 > 0 {
			temp = -temp
		}
	default:
		// keeping this for retro-compatibility but not tested
		hum = binary.LittleEndian.Uint16(buf[0:2])
		temp = int16(buf[3])<<8 + int16(buf[2]&0x7f)
		if buf[2]&0x80 > 0 {
			temp = -temp
		}
	}
	return
}
