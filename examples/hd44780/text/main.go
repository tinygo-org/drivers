package main

import (
	"machine"

	"tinygo.org/x/drivers/hd44780"
)

func main() {

	lcd, _ := hd44780.NewGPIO4Bit(
		[]machine.Pin{machine.P0, machine.P1, machine.P2, machine.P3},
		machine.P4,
		machine.P5,
		machine.P6,
	)

	lcd.Configure(hd44780.Config{
		Width:       16,
		Height:      2,
		CursorOnOff: true,
		CursorBlink: true,
	})

	lcd.Write([]byte("This is a long line"))
	lcd.Display()

	for {

	}
}
