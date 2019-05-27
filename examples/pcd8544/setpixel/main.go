package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/pcd8544"
)

func main() {
	dcPin := machine.P3
	dcPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rstPin := machine.P4
	rstPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	scePin := machine.P5
	scePin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	machine.SPI0.Configure(machine.SPIConfig{})

	lcd := pcd8544.New(machine.SPI0, dcPin, rstPin, scePin)
	lcd.Configure(pcd8544.Config{})

	var x int16
	var y int16
	deltaX := int16(1)
	deltaY := int16(1)
	for {
		pixel := lcd.GetPixel(x, y)
		c := color.RGBA{255, 255, 255, 255}
		if pixel {
			c = color.RGBA{0, 0, 0, 255}
		}
		lcd.SetPixel(x, y, c)
		lcd.Display()

		x += deltaX
		y += deltaY

		if x == 0 || x == 83 {
			deltaX = -deltaX
		}

		if y == 0 || y == 47 {
			deltaY = -deltaY
		}
		time.Sleep(1 * time.Millisecond)
	}
}
