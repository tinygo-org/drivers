package main

import (
	"machine"

	"image/color"

	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ledstripdisplay"
	"tinygo.org/x/drivers/ws2812"
)

func main() {

	ledPin := machine.A1
	ledPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	ledstrip := ws2812.New(ledPin)

	width := int16(8)
	height := int16(8)
	display := ledstripdisplay.New(ledstrip, width, height, ledstripdisplay.PARALLEL_1)
	display.Configure(ledstripdisplay.Config{Rotation: ledstripdisplay.NO_ROTATION})

	white := color.RGBA{255, 255, 255, 255}
	red := color.RGBA{255, 0, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	green := color.RGBA{0, 255, 0, 255}
	black := color.RGBA{0, 0, 0, 255}

	drawFilledRectangle(&display, 0, 0, width/2, height/2, white)
	drawFilledRectangle(&display, width/2, 0, width/2, height/2, red)
	drawFilledRectangle(&display, 0, height/2, width/2, height/2, green)
	drawFilledRectangle(&display, width/2, height/2, width/2, height/2, blue)
	drawFilledRectangle(&display, width/4, height/4, width/2, height/2, black)
	for {
		display.Display()

		time.Sleep(10000 * time.Millisecond)
	}
}

func drawFilledRectangle(display drivers.Displayer, x, y, width, height int16, c color.RGBA) {
	for i := int16(0); i < width; i++ {
		for j := int16(0); j < height; j++ {
			println(x+i, y+j, width, height)
			display.SetPixel(x+i, y+j, c)
		}
	}
}
