package main

import (
	"fmt"
	"machine"

	"github.com/tinygo-org/drivers/gps"
)

func main() {
	println("GPS UART Example")
	machine.UART1.Configure(machine.UARTConfig{BaudRate: 9600})
	ublox := gps.NewUART(&machine.UART1)
	parser := gps.Parser(ublox)
	var fix gps.Fix
	for {
		fix = parser.NextFix()
		if fix.Valid {
			print(fix.Time.Format("15:04:05"))
			print(", lat=", fmt.Sprintf("%f", fix.Latitude))
			print(", long=", fmt.Sprintf("%f", fix.Longitude))
			print(", altitude:=", fix.Altitude)
			print(", satellites=", fix.Satellites)
			println()
		} else {
			println("No fix")
		}
	}
}
