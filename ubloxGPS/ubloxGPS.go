// Package ubloxGPS provides a driver for UBlox GPS receivers over I2C
//
// Datasheet:
// https://www.u-blox.com/sites/default/files/products/documents/u-blox8-M8_ReceiverDescrProtSpec_%28UBX-13003221%29_Public.pdf
// (Section 11.5)
//
package ubloxGPS

import (
	"fmt"
	"machine"
	"strconv"
	"strings"
	"time"
)

// Device wraps an I2C connection to a ublox gps device.
type Device struct {
	bus     machine.I2C
	Address uint16
	buff    *[2048]byte
}

// New creates a new GPS connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not initialize the device.
// You must call Configure() first in order to use the device itself.
func New(bus machine.I2C) Device {
	return Device{
		bus:     bus,
		Address: Address,
		buff:    new([2048]byte),
	}
}

// Available returns how many bytes of GPS data are currently available.
func (d *Device) Available() (available int) {
	// println("Available")
	dataLen := []byte{0, 0}
	d.bus.ReadRegister(uint8(d.Address), FD, dataLen)
	available = int(dataLen[0])*256 + int(dataLen[1])
	return available
}

func (d *Device) Read() (result string, err error) {
	// println("Read")
	var available = d.Available()
	if available < 1 {
		return "", err
	}
	data := d.buff[0:available]
	d.bus.ReadRegister(uint8(d.Address), FF, data)
	result = string(data)
	return result, err
}

func (d *Device) ReadSentences() (result []string, err error) {
	// println("ReadSentences")
	var s, _ = d.Read()
	result = strings.Split(s, "\r\n")
	return result, err
}

// func validSentence(s string) bool {
// 	if len(s) > 0 && s[0] == '$' {
// 		var end = len(s)
// 		if s[end-3] == '*' {
// 			return true
// 		}
// 	}
// 	return false
// }
//
// func remove(s []string, i int) []string {
// 	s[i] = s[len(s)-1]
// 	// We do not need to put s[i] at the end, as it will be discarded anyway
// 	return s[:len(s)-1]
// }

func (d *Device) ReadSentence(stype string) (result string) {
	// println("ReadSentence")
	var sentences, _ = d.ReadSentences()
	for _, s := range sentences {
		if len(s) > 6 && s[0] == '$' {
			var end = len(s)
			if s[end-3] == '*' {
				if s[3:6] == stype {
					return s
				}
			}
		}
	}
	return ""
}

// fix is a GPS location fix
type Fix struct {
	Valid     bool
	Time      string
	Latitude  string
	Longitude string
	Altitude  int32
	Satelites int16
}

func (d *Device) ReadFix() (result Fix) {
	// println("ReadFix")
	result = Fix{Valid: false}
	var ggaSentence = d.readGGA()
	if len(ggaSentence) < 1 {
		return result
	}
	println(ggaSentence)
	var ggaFields = strings.Split(ggaSentence, ",")
	result.Valid = true
	result.Altitude = findAltitude(ggaFields)
	result.Satelites = findSatelites(ggaFields)
	result.Longitude = findLongitude(ggaFields)
	result.Latitude = findLatitude(ggaFields)
	result.Time = findTime(ggaFields)
	return result
}

func (d *Device) readGGA() (gga string) {
	// println("ReadGGA")
	// $--GGA,,,,,,,,,,,,,,*hh
	for i := 1; i <= 5; i++ {
		var sentence = d.ReadSentence("GGA")
		if len(sentence) > 0 {
			return sentence
		}
		time.Sleep(500 * time.Millisecond)
	}
	return ""
}

func findTime(ggaFields []string) (t string) {
	// println("findTime")
	// $GNGGA,hhmmss.ss,,,,,,,,,,,,,*63
	if len(ggaFields) < 1 && len(ggaFields[1]) < 6 {
		return "hh:mm:ss"
	}
	hh := string(ggaFields[1][0:2])
	mm := string(ggaFields[1][2:4])
	ss := string(ggaFields[1][4:6])
	return hh + ":" + mm + ":" + ss
}

func findAltitude(ggaFields []string) (a int32) {
	// println("findAltitude")
	// $GNGGA,,,,,,,,,25.8,,,,,*63
	if len(ggaFields) > 8 && len(ggaFields[9]) > 0 {
		var v, _ = strconv.ParseFloat(ggaFields[9], 32)
		return int32(v)
	}
	return -99999
}

func findLatitude(ggaFields []string) (l string) {
	// println("findLatitude")
	// $--GGA,,ddmm.mmmmm,x,,,,,,,,,,,*hh
	if len(ggaFields) > 2 && len(ggaFields[2]) > 8 {
		var dd = ggaFields[2][0:2]
		var mm = ggaFields[2][2:]
		var d, _ = strconv.ParseFloat(dd, 32)
		var m, _ = strconv.ParseFloat(mm, 32)
		var v = (d + (m / 60))
		if ggaFields[3] == "S" {
			v *= -1
		}
		return fmt.Sprintf("%f", v)
	}
	return "-0.0"
}

func findLongitude(ggaFields []string) (l string) {
	// println("findLongitude")
	// $--GGA,,,,dddmm.mmmmm,x,,,,,,,,,*hh
	if len(ggaFields) > 4 && len(ggaFields[4]) > 8 {
		var ddd = ggaFields[4][0:3]
		var mm = ggaFields[4][3:]
		var d, _ = strconv.ParseFloat(ddd, 32)
		var m, _ = strconv.ParseFloat(mm, 32)
		var v = (d + (m / 60))
		if ggaFields[5] == "W" {
			v *= -1
		}
		var s = fmt.Sprintf("%f", v)
		return s
	}
	return "-0.0"
}

func findSatelites(ggaFields []string) (n int16) {
	// println("findSatelites")
	// $--GGA,,,,,,,nn,,,,,,,*hh
	if len(ggaFields) > 6 && len(ggaFields[7]) > 0 {
		var nn = ggaFields[7]
		var v, _ = strconv.ParseInt(nn, 10, 32)
		n = int16(v)
		return n
	}
	return 0
}
