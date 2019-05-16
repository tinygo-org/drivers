package main

import (
	"machine"

	"github.com/tinygo-org/drivers/gps"
)

func main() {
	println("GPS Example")
	machine.I2C0.Configure(machine.I2CConfig{})
	ublox := gps.NewI2C(machine.I2C0)
	parser := gps.Parser(ublox)
	var fix gps.Fix
	for {
		fix = parser.NextFix()
		if fix.Valid {
			print(fix.Time)
			print(", lat=", fix.Latitude)
			print(", long=", fix.Longitude)
			print(", altitude:=", fix.Altitude)
			print(", satellites=", fix.Satellites)
			println()
		} else {
			println("No fix")
		}
	}
}
