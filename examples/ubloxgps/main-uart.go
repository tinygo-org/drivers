package main

import (
	"machine"

	"github.com/tinygo-org/drivers/gps"
)

var (
	uart = machine.UART1
)

func main() {
	println("GPS UART Example")
	machine.UART1.Configure(machine.UARTConfig{BaudRate: 9600})
	ublox := gps.New(&machine.UART1)
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
