package main

import (
	"machine"

	"github.com/tinygo-org/drivers/ubloxGPS"
)

func main() {
	println("GPS Example")
	machine.I2C0.Configure(machine.I2CConfig{})
	gps := ubloxGPS.New(machine.I2C0)
	var f ubloxGPS.Fix
	for {
		f = gps.ReadFix()
		if f.Valid {
			print(f.Time)
			print(", lat=", f.Latitude)
			print(", long=", f.Longitude)
			print(", altitude:=", f.Altitude)
			print(", satelites=", f.Satelites)
			println()
		} else {
			println("No fix")
		}
	}
}
