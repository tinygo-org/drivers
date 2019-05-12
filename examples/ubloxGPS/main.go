package main

import (
	"machine"

	"github.com/tinygo-org/drivers/ubloxGPS"
)

func main() {
	println("GPS Example2")
	machine.I2C0.Configure(machine.I2CConfig{})
	gps := ubloxGPS.New(machine.I2C0)
	parser := ubloxGPS.Parser(gps)
	var fix ubloxGPS.Fix
	for {
		parser.NextFix(&fix)
		if fix.Valid {
			print(fix.Time)
			print(", lat=", fix.Latitude)
			print(", long=", fix.Longitude)
			print(", altitude:=", fix.Altitude)
			print(", satelites=", fix.Satelites)
			println()
		} else {
			println("No fix")
		}
	}
}
