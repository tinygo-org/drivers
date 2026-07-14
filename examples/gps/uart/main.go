package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/gps"
)

func main() {
	machine.UART1.Configure(machine.UARTConfig{BaudRate: 9600})
	ublox := gps.NewUART(machine.UART1)
	parser := gps.NewParser()
	var fix gps.Fix
	for {
		s, err := ublox.NextSentence()
		if err != nil {
			switch err {
			case gps.ErrUnknownNMEASentence, gps.ErrInvalidNMEASentence, gps.ErrInvalidNMEASentenceLength:
				continue
			default:
				println("sentence error:", err)
				continue
			}
		}

		fix, err = parser.Parse(s)
		if err != nil {
			switch err {
			case gps.ErrUnknownNMEASentence, gps.ErrInvalidNMEASentence, gps.ErrInvalidNMEASentenceLength:
				continue
			default:
				println("parse error:", err)
				continue
			}
		}
		if fix.Valid {
			print(fix.Time.Format("15:04:05"))
			print(", lat=")
			print(fix.Latitude)
			print(", long=")
			print(fix.Longitude)
			print(", altitude=", fix.Altitude)
			print(", satellites=", fix.Satellites)
			if fix.Speed != 0 {
				print(", speed=")
				print(fix.Speed)
			}
			if fix.Heading != 0 {
				print(", heading=")
				print(fix.Heading)
			}
			println()
		} else {
			if fix.Type == gps.GSV {
				// GSV sentence provides satellite count even if no fix yet
				println(fix.Satellites, "satellites visible")
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
}
