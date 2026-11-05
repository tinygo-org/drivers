// Demo for the Freetronics 16x2 LCD and keypad shield on an Arduino Uno.
// Pinout from https://www.shieldlist.org/freetronics/16x2lcd-keypad
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/hd44780"
)

func main() {
	backlight := machine.D3
	backlight.Configure(machine.PinConfig{Mode: machine.PinOutput})
	backlight.High()

	lcd, err := hd44780.NewGPIO4Bit(
		[]machine.Pin{machine.D4, machine.D5, machine.D6, machine.D7},
		machine.D9,
		machine.D8,
		machine.NoPin,
	)
	if err != nil {
		println(err.Error())
		return
	}
	lcd.Configure(hd44780.Config{
		Width:  16,
		Height: 2,
	})

	machine.InitADC()
	buttons := machine.ADC{Pin: machine.ADC0}
	buttons.Configure(machine.ADCConfig{})

	show(&lcd, 0, "TinyGo 1602")
	last := ""
	for {
		name := button(buttons.Get() >> 6)
		if name != last {
			show(&lcd, 1, name)
			last = name
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// button decodes the A0 resistor ladder. Values are 10 bit.
func button(v uint16) string {
	switch {
	case v < 50:
		return "RIGHT"
	case v < 200:
		return "UP"
	case v < 400:
		return "DOWN"
	case v < 600:
		return "LEFT"
	case v < 850:
		return "SELECT"
	default:
		return "-"
	}
}

func show(lcd *hd44780.Device, row uint8, s string) {
	line := []byte("                ")
	copy(line, s)
	lcd.SetCursor(0, row)
	lcd.Write(line)
	lcd.Display()
}
