package gps

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

var (
	errEmptyNMEASentence   = errors.New("cannot parse empty NMEA sentence")
	errRMCNotSupportedYet  = errors.New("RMC NMEA sentence not yet supported")
	errGSANotSupportedYet  = errors.New("GSA NMEA sentence not yet supported")
	errGSVNotSupportedYet  = errors.New("GSV NMEA sentence not yet supported")
	errUnknownNMEASentence = errors.New("unknown NMEA sentence type")
)

// Parser for GPS NMEA sentences.
type Parser struct {
}

// Fix is a GPS location fix
type Fix struct {
	Valid      bool
	Time       time.Time
	Latitude   float32
	Longitude  float32
	Altitude   int32
	Satellites int16
}

// NewParser returns a GPS NMEA Parser.
func NewParser() Parser {
	return Parser{}
}

// Parse parses a NMEA sentence looking for fix info.
func (parser *Parser) Parse(sentence string) (fix Fix, err error) {
	if sentence == "" {
		err = errEmptyNMEASentence
		return
	}
	typ := sentence[3:6]
	switch typ {
	case "GGA":
		var ggaFields = strings.Split(sentence, ",")
		fix.Altitude = findAltitude(ggaFields)
		fix.Satellites = findSatellites(ggaFields)
		fix.Longitude = findLongitude(ggaFields)
		fix.Latitude = findLatitude(ggaFields)
		fix.Time = findTime(ggaFields)
		fix.Valid = (fix.Altitude != -99999) && (fix.Satellites > 0)
	case "RMC":
		err = errRMCNotSupportedYet
	case "GSA":
		err = errGSANotSupportedYet
	case "GSV":
		err = errGSVNotSupportedYet
	default:
		err = errUnknownNMEASentence
	}
	return
}

// findTime returns the time from a GGA sentence:
// $--GGA,hhmmss.ss,,,,,,,,,,,,,*xx
func findTime(ggaFields []string) time.Time {
	if len(ggaFields) < 1 || len(ggaFields[1]) < 6 {
		return time.Time{}
	}

	h, _ := strconv.ParseInt(ggaFields[1][0:2], 10, 8)
	m, _ := strconv.ParseInt(ggaFields[1][2:4], 10, 8)
	s, _ := strconv.ParseInt(ggaFields[1][4:6], 10, 8)
	ms, _ := strconv.ParseInt(ggaFields[1][7:10], 10, 16)
	t := time.Date(0, 0, 0, int(h), int(m), int(s), int(ms), time.UTC)

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
