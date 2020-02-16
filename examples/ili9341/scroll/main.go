package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/ili9341"
)

var (
	display = ili9341.NewParallel(
		machine.LCD_DATA0,
		machine.TFT_WR,
		machine.TFT_DC,
		machine.TFT_CS,
		machine.TFT_RESET,
		machine.TFT_RD,
	)

	red   = color.RGBA{255, 0, 0, 255}
	blue  = color.RGBA{0, 0, 255, 255}
	green = color.RGBA{0, 255, 0, 255}
	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}
)

func main() {

	machine.TFT_BACKLIGHT.Configure(machine.PinConfig{machine.PinOutput})

	display.Configure(ili9341.Config{})
	width, height := display.Size()

	display.FillScreen(black)
	machine.TFT_BACKLIGHT.High()

	display.FillRectangle(0, 0, width/2, height/2, white)
	display.FillRectangle(width/2, 0, width/2, height/2, red)
	display.FillRectangle(0, height/2, width/2, height/2, green)
	display.FillRectangle(width/2, height/2, width/2, height/2, blue)
	display.FillRectangle(width/4, height/4, width/2, height/2, black)

	for scroll := int16(0); ; scroll = (scroll + 1) % 320 {
		time.Sleep(7500 * time.Microsecond)
		display.SetScroll(scroll)
	}

}
