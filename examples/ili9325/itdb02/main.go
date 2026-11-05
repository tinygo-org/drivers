// Draws a test pattern on the ITead ITDB02-2.8 shield in landscape, with the
// red square in the top left corner.
package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ili9325"
)

var (
	data = [8]machine.Pin{machine.D0, machine.D1, machine.D2, machine.D3, machine.D4, machine.D5, machine.D6, machine.D7}
	rs   = machine.ADC5
	wr   = machine.ADC4
	cs   = machine.ADC3
	rst  = machine.ADC2
)

var (
	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}
	red   = color.RGBA{255, 0, 0, 255}
	green = color.RGBA{0, 255, 0, 255}
	blue  = color.RGBA{0, 0, 255, 255}
)

func main() {
	out := machine.PinConfig{Mode: machine.PinOutput}
	for _, p := range append(data[:], rs, wr, cs, rst) {
		p.Configure(out)
	}

	display := ili9325.New(newBus(), rs, cs, rst)
	display.Configure(ili9325.Config{Rotation: drivers.Rotation90})

	w, h := display.Size()
	display.FillScreen(black)
	display.FillRectangle(0, 0, 40, 40, red)
	display.FillRectangle(w-40, 0, 40, 40, green)
	display.FillRectangle(0, h-40, 40, 40, blue)
	display.FillRectangle(w-40, h-40, 40, 40, white)

	// Only redraw the strips that change, so the square doesn't flicker.
	x, y, dx := int16(40), h/2-20, int16(4)
	display.FillRectangle(x, y, 40, 40, green)
	for {
		if x+dx < 40 || x+dx > w-80 {
			dx = -dx
		}
		if dx > 0 {
			display.FillRectangle(x, y, dx, 40, black)
			display.FillRectangle(x+40, y, dx, 40, green)
		} else {
			display.FillRectangle(x+40+dx, y, -dx, 40, black)
			display.FillRectangle(x+dx, y, -dx, 40, green)
		}
		x += dx
		time.Sleep(10 * time.Millisecond)
	}
}
