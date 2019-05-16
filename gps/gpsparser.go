package gps

import (
	"fmt"
	"strconv"
	"strings"
)

type GPSParser struct {
	gpsDevice GPSDevice
}

// fix is a GPS location fix
type Fix struct {
	Valid      bool
	Time       string
	Latitude   string
	Longitude  string
	Altitude   int32
	Satellites int16
}

func Parser(gpsDevice GPSDevice) GPSParser {
	return GPSParser{
		gpsDevice: gpsDevice,
	}
}

func (parser *GPSParser) NextFix() (fix Fix) {
	var ggaSentence = nextGGA(parser.gpsDevice)
	var ggaFields = strings.Split(ggaSentence, ",")
	fix.Valid = true
	fix.Altitude = findAltitude(ggaFields)
	fix.Satellites = findSatellites(ggaFields)
	fix.Longitude = findLongitude(ggaFields)
	fix.Latitude = findLatitude(ggaFields)
	fix.Time = findTime(ggaFields)
	return fix
}

func nextGGA(gpsDevice GPSDevice) (sentence string) {
	// $--GGA,,,,,,,,,,,,,,*hh
	for {
		sentence = gpsDevice.ReadNextSentence()
		if sentence[3:6] == "GGA" {
			return sentence
		}
	}
}

func findTime(ggaFields []string) (t string) {
	// $GNGGA,hhmmss.ss,,,,,,,,,,,,,*63
	if len(ggaFields) < 1 || len(ggaFields[1]) < 6 {
		return "hh:mm:ss"
	}
	hh := string(ggaFields[1][0:2])
	mm := string(ggaFields[1][2:4])
	ss := string(ggaFields[1][4:6])
	return hh + ":" + mm + ":" + ss
}

func findAltitude(ggaFields []string) (a int32) {
	// $GNGGA,,,,,,,,,25.8,,,,,*63
	if len(ggaFields) > 8 && len(ggaFields[9]) > 0 {
		var v, _ = strconv.ParseFloat(ggaFields[9], 32)
		return int32(v)
	}
	return -99999
}

func findLatitude(ggaFields []string) (l string) {
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

func findSatellites(ggaFields []string) (n int16) {
	// $--GGA,,,,,,,nn,,,,,,,*hh
	if len(ggaFields) > 6 && len(ggaFields[7]) > 0 {
		var nn = ggaFields[7]
		var v, _ = strconv.ParseInt(nn, 10, 32)
		n = int16(v)
		return n
	}
	return 0
}
