package gps

import (
	"strconv"
	"strings"
	"time"
)

// Parser for GPS NMEA sentences.
type Parser struct {
}

// Fix is a GPS location fix
type Fix struct {
	// Valid if the fix was valid.
	Valid bool

	// Time that the fix was taken, in UTC time.
	Time time.Time

	// Latitude is the decimal latitude. Negative numbers indicate S.
	Latitude float32

	// Longitude is the decimal longitude. Negative numbers indicate E.
	Longitude float32

	// Altitude is only returned for GGA sentences.
	Altitude int32

	// Satellites is the number of visible satellites, but is only returned for GGA sentences.
	Satellites int16

	// Speed based on reported movement. Only returned for RMC sentences.
	Speed float32

	// Heading based on reported movement. Only returned for RMC sentences.
	Heading float32
}

// NewParser returns a GPS NMEA Parser.
func NewParser() Parser {
	return Parser{}
}

// Parse parses a NMEA sentence looking for fix info.
func (parser *Parser) Parse(sentence string) (Fix, error) {
	var fix Fix
	if sentence == "" {
		return fix, errEmptyNMEASentence
	}
	if len(sentence) < 6 {
		return fix, errInvalidNMEASentenceLength
	}
	typ := sentence[3:6]
	switch typ {
	case "GGA":
		// https://docs.novatel.com/OEM7/Content/Logs/GPGGA.htm
		fields := strings.Split(sentence, ",")
		if len(fields) != 15 {
			return fix, errInvalidGGASentence
		}

		fix.Time = findTime(fields[1])
		fix.Latitude = findLatitude(fields[2], fields[3])
		fix.Longitude = findLongitude(fields[4], fields[5])
		fix.Satellites = findSatellites(fields[7])
		fix.Altitude = findAltitude(fields[9])
		fix.Valid = (fix.Altitude != -99999) && (fix.Satellites > 0)

		return fix, nil
	case "GLL":
		// https://docs.novatel.com/OEM7/Content/Logs/GPGLL.htm
		fields := strings.Split(sentence, ",")
		if len(fields) != 8 {
			return fix, errInvalidGLLSentence
		}

		fix.Latitude = findLatitude(fields[1], fields[2])
		fix.Longitude = findLongitude(fields[3], fields[4])
		fix.Time = findTime(fields[5])

		fix.Valid = (fields[6] == "A")

		return fix, nil
	case "RMC":
		// https://docs.novatel.com/OEM7/Content/Logs/GPRMC.htm
		fields := strings.Split(sentence, ",")
		if len(fields) != 13 {
			return fix, errInvalidRMCSentence
		}

		fix.Time = findTime(fields[1])
		fix.Valid = (fields[2] == "A")
		fix.Latitude = findLatitude(fields[3], fields[4])
		fix.Longitude = findLongitude(fields[5], fields[6])
		fix.Speed = findSpeed(fields[7])
		fix.Heading = findHeading(fields[8])
		date := findDate(fields[9])
		fix.Time = fix.Time.AddDate(date.Year(), int(date.Month()), date.Day())

		return fix, nil
	}

	return fix, newGPSError(errUnknownNMEASentence, sentence, typ)
}

// findTime returns the time from an NMEA sentence:
// $--GGA,hhmmss.ss,,,,,,,,,,,,,*xx
func findTime(val string) time.Time {
	if len(val) < 6 {
		return time.Time{}
	}

	h, _ := strconv.ParseInt(val[0:2], 10, 8)
	m, _ := strconv.ParseInt(val[2:4], 10, 8)
	s, _ := strconv.ParseInt(val[4:6], 10, 8)
	ms := int64(0)
	if len(val) > 6 {
		ms, _ = strconv.ParseInt(val[7:], 10, 16)
	}
	t := time.Date(0, 0, 0, int(h), int(m), int(s), int(ms), time.UTC)

	return t
}

// findAltitude returns the altitude from an NMEA sentence:
// $--GGA,,,,,,,,,25.8,,,,,*63
func findAltitude(val string) int32 {
	if len(val) > 0 {
		var v, _ = strconv.ParseFloat(val, 32)
		return int32(v)
	}
	return -99999
}

// findLatitude returns the Latitude from an NMEA sentence:
// $--GGA,,ddmm.mmmmm,x,,,,,,,,,,,*hh
func findLatitude(val, hemi string) float32 {
	if len(val) > 8 {
		dd := val[0:2]
		mm := val[2:]
		d, _ := strconv.ParseFloat(dd, 32)
		m, _ := strconv.ParseFloat(mm, 32)
		v := float32(d + (m / 60))
		if hemi == "S" {
			v *= -1
		}
		return v
	}
	return 0.0
}

// findLongitude returns the longitude from an NMEA sentence:
// $--GGA,,,,dddmm.mmmmm,x,,,,,,,,,*hh
func findLongitude(val, hemi string) float32 {
	if len(val) > 8 {
		var ddd = val[0:3]
		var mm = val[3:]
		var d, _ = strconv.ParseFloat(ddd, 32)
		var m, _ = strconv.ParseFloat(mm, 32)
		var v = float32(d + (m / 60))
		if hemi == "W" {
			v *= -1
		}
		return v
	}
	return 0.0
}

// findSatellites returns the satellites from an NMEA sentence:
// $--GGA,,,,,,,nn,,,,,,,*hh
func findSatellites(val string) (n int16) {
	if len(val) > 0 {
		var nn = val
		var v, _ = strconv.ParseInt(nn, 10, 32)
		n = int16(v)
		return n
	}
	return 0
}

// findDate returns the date from an RMC NMEA sentence.
func findDate(val string) time.Time {
	if len(val) < 6 {
		return time.Time{}
	}

	d, _ := strconv.ParseInt(val[0:2], 10, 8)
	m, _ := strconv.ParseInt(val[2:4], 10, 8)
	y, _ := strconv.ParseInt(val[4:6], 10, 8)
	t := time.Date(int(2000+y), time.Month(m), int(d), 0, 0, 0, 0, time.UTC)

	return t
}

// findSpeed returns the speed from an RMC NMEA sentence.
func findSpeed(val string) float32 {
	if len(val) > 0 {
		var v, _ = strconv.ParseFloat(val, 32)
		return float32(v)
	}
	return 0
}

// findHeading returns the speed from an RMC NMEA sentence.
func findHeading(val string) float32 {
	if len(val) > 0 {
		var v, _ = strconv.ParseFloat(val, 32)
		return float32(v)
	}
	return 0
}
