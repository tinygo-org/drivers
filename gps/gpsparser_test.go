package gps

import (
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

func TestParseUnknownSentence(t *testing.T) {
	c := qt.New(t)

	p := NewParser()

	val := "$GPGSV,3,1,09,07,14,317,22,08,31,284,25,10,32,133,39,16,85,232,29*7F"
	_, err := p.Parse(val)
	c.Assert(err.Error(), qt.Contains, "unsupported NMEA sentence type")
}

func TestParseGGA(t *testing.T) {
	c := qt.New(t)

	p := NewParser()

	val := "$GPGGA,115739.00,4158.8441367,N,09147.4416929,"
	fix, err := p.Parse(val)
	if err != errInvalidGGASentence {
		t.Error("should have errInvalidGGASentence error")
	}

	val = "$GPGGA,115739.00,4158.8441367,N,09147.4416929,W,4,13,0.9,255.747,M,-32.00,M,01,0000*6E"
	fix, err = p.Parse(val)
	if err != nil {
		t.Error("should have parsed")
	}
	c.Assert(fix.Latitude, qt.Equals, float32(41.980735778808594))
	c.Assert(fix.Longitude, qt.Equals, float32(-91.79069519042969))
	c.Assert(fix.Altitude, qt.Equals, int32(255))
}

func TestParseGLL(t *testing.T) {
	c := qt.New(t)

	p := NewParser()

	val := "$GPGLL,3953.88008971,N,10506.7531891"
	_, err := p.Parse(val)
	if err != errInvalidGLLSentence {
		t.Error("should have errInvalidGLLSentence error")
	}

	val = "$GPGLL,5109.0262317,N,11401.8407304,W,202725.00,A,D*79"
	fix, err := p.Parse(val)
	if err != nil {
		t.Error("should have parsed")
	}

	c.Assert(fix.Latitude, qt.Equals, float32(51.15043640136719))
	c.Assert(fix.Longitude, qt.Equals, float32(-114.03067779541016))
}

func TestParseRMC(t *testing.T) {
	c := qt.New(t)

	p := NewParser()

	val := "$GPRMC,203522.00,A,5109.0262308,N,11401.8407342,"
	_, err := p.Parse(val)
	if err != errInvalidRMCSentence {
		t.Error("should have errInvalidRMCSentence error")
	}

	val = "$GPRMC,203522.00,A,5109.0262308,N,11401.8407342,W,0.004,133.4,130522,0.0,E,D*2B"
	fix, err := p.Parse(val)
	if err != nil {
		t.Error("should have parsed")
	}

	c.Assert(fix.Time.Year(), qt.Equals, 2022)
	c.Assert(fix.Time.Month(), qt.Equals, time.May)
	c.Assert(fix.Time.Day(), qt.Equals, 13)
	c.Assert(fix.Time.Hour(), qt.Equals, 20)
	c.Assert(fix.Time.Minute(), qt.Equals, 35)
	c.Assert(fix.Time.Second(), qt.Equals, 22)
	c.Assert(fix.Latitude, qt.Equals, float32(51.15043640136719))
	c.Assert(fix.Longitude, qt.Equals, float32(-114.03067779541016))
}

func TestTime(t *testing.T) {
	c := qt.New(t)

	val := ""
	tm := findTime(val)
	c.Assert(tm, qt.Equals, time.Time{})

	val = "225446"
	tm = findTime(val)
	c.Assert(tm, qt.Equals, time.Date(0, 0, 0, 22, 54, 46, 0, time.UTC))

	val = "124326.02752"
	tm = findTime(val)
	c.Assert(tm, qt.Equals, time.Date(0, 0, 0, 12, 43, 26, 2752, time.UTC))
}
