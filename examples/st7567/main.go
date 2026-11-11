// Draws a border, a cross and a block in the top left corner on an ST7567 LCD,
// such as the WeAct Studio 1.37 inch Matrix LCD module.
package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/st7567"
)

func main() {
	for _, p := range []machine.Pin{blPin, rstPin, dcPin, csPin} {
		p.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
	blPin.High()

	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8 * machine.MHz,
		SCK:       sckPin,
		SDO:       sdoPin,
		SDI:       machine.NoPin,
	})

	display := st7567.New(machine.SPI0, csPin, dcPin, rstPin)
	if err := display.Configure(st7567.Config{}); err != nil {
		for {
			println("st7567:", err.Error())
			time.Sleep(time.Second)
		}
	}

	on := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	w, h := display.Size()
	for x := int16(0); x < w; x++ {
		display.SetPixel(x, 0, on)
		display.SetPixel(x, h-1, on)
	}
	for y := int16(0); y < h; y++ {
		display.SetPixel(0, y, on)
		display.SetPixel(w-1, y, on)
		display.SetPixel(y*w/h, y, on)
		display.SetPixel(w-1-y*w/h, y, on)
	}
	for y := int16(2); y < 12; y++ {
		for x := int16(2); x < 22; x++ {
			display.SetPixel(x, y, on)
		}
	}
	display.Display()

	for inverted := false; ; inverted = !inverted {
		display.SetInverted(inverted)
		time.Sleep(2 * time.Second)
	}
}
