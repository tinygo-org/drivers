package main

import (
	"image/color"
	"machine"

	"tinygo.org/x/drivers/waveshare-epd/epd1in54"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/gophers"
)

var (
	spi0 = machine.SPI0
	cs   = machine.D10
	dc   = machine.D9
	rst  = machine.D6
	busy = machine.D5

	black = color.RGBA{R: 1, G: 1, B: 1, A: 255}
)

func main() {
	display := epd1in54.New(spi0, cs, dc, rst, busy)

	display.LDirInit(epd1in54.Config{})
	display.Clear()
	display.ClearBuffer()

	tinyfont.WriteLineRotated(&display, &gophers.Regular58pt, 150, 0, "A B C", black, tinyfont.ROTATION_90)
	tinyfont.WriteLineRotated(&display, &gophers.Regular58pt, 100, 0, "D E F", black, tinyfont.ROTATION_90)
	tinyfont.WriteLineRotated(&display, &gophers.Regular58pt, 50, 0, "G H I", black, tinyfont.ROTATION_90)
	tinyfont.WriteLineRotated(&display, &gophers.Regular58pt, 0, 0, "J K L", black, tinyfont.ROTATION_90)

	display.Display()
	display.Sleep()
}
