package gps

import (
	"strconv"
	"strings"
	"time"
)

type GPSParser struct {
	gpsDevice GPSDevice
}

// fix is a GPS location fix
type Fix struct {
	Valid      bool
	Time       time.Time
	Latitude   float32
	Longitude  float32
	Altitude   int32
	Satellites int16
}

func Parser(gpsDevice GPSDevice) GPSParser {
	return GPSParser{
		gpsDevice: gpsDevice,
	}
}

// NextFix returns the next GPS location Fix from the GPS device
func (parser *GPSParser) NextFix() (fix Fix) {
	var ggaSentence = nextGGA(parser.gpsDevice)
	var ggaFields = strings.Split(ggaSentence, ",")
	fix.Altitude = findAltitude(ggaFields)
	fix.Satellites = findSatellites(ggaFields)
	fix.Longitude = findLongitude(ggaFields)
	fix.Latitude = findLatitude(ggaFields)
	fix.Time = findTime(ggaFields)
	fix.Valid = (fix.Altitude != -99999) && (fix.Satellites > 0)
	return fix
}

// nextGGA returns the next GGA type sentence from the GPS device
// $--GGA,,,,,,,,,,,,,,*hh
func nextGGA(gpsDevice GPSDevice) (sentence string) {
	for {
		sentence = gpsDevice.ReadNextSentence()
		if sentence[3:6] == "GGA" {
			return sentence
		}
	}
}

// findTime returns the time from a GGA sentence:
// $--GGA,hhmmss.ss,,,,,,,,,,,,,*xx
func findTime(ggaFields []string) time.Time {
	if len(ggaFields) < 1 || len(ggaFields[1]) < 6 {
		return time.Time{}
	}
	ts := strings.Builder{}
	ts.WriteString(ggaFields[1][0:2])
	ts.WriteString(":")
	ts.WriteString(ggaFields[1][2:4])
	ts.WriteString(":")
	ts.WriteString(ggaFields[1][4:6])
	var t, _ = time.Parse("15:04:05", ts.String())

	return t
}

// findAltitude returns the altitude from a GGA sentence:
// $--GGA,,,,,,,,,25.8,,,,,*63
func findAltitude(ggaFields []string) int32 {
	if len(ggaFields) > 8 && len(ggaFields[9]) > 0 {
		var v, _ = strconv.ParseFloat(ggaFields[9], 32)
		return int32(v)
	}
	return -99999
}

// findLatitude returns the Latitude from a GGA sentence:
// $--GGA,,ddmm.mmmmm,x,,,,,,,,,,,*hh
func findLatitude(ggaFields []string) float32 {
	if len(ggaFields) > 2 && len(ggaFields[2]) > 8 {
		var dd = ggaFields[2][0:2]
		var mm = ggaFields[2][2:]
		var d, _ = strconv.ParseFloat(dd, 32)
		var m, _ = strconv.ParseFloat(mm, 32)
		var v = float32(d + (m / 60))
		if ggaFields[3] == "S" {
			v *= -1
		}
		return v
	}
	return 0.0
}

// findLatitude returns the longitude from a GGA sentence:
// $--GGA,,,,dddmm.mmmmm,x,,,,,,,,,*hh
func findLongitude(ggaFields []string) float32 {
	if len(ggaFields) > 4 && len(ggaFields[4]) > 8 {
		var ddd = ggaFields[4][0:3]
		var mm = ggaFields[4][3:]
		var d, _ = strconv.ParseFloat(ddd, 32)
		var m, _ = strconv.ParseFloat(mm, 32)
		var v = float32(d + (m / 60))
		if ggaFields[5] == "W" {
			v *= -1
		}
		return v
	}
	return 0.0
}

// findSatellites returns the satellites from a GGA sentence:
// $--GGA,,,,,,,nn,,,,,,,*hh
func findSatellites(ggaFields []string) (n int16) {
	if len(ggaFields) > 6 && len(ggaFields[7]) > 0 {
		var nn = ggaFields[7]
		var v, _ = strconv.ParseInt(nn, 10, 32)
		n = int16(v)
		return n
	}
	return 0
}
