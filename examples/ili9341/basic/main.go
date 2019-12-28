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
	/*
		display = ili9341.NewSPI(
			&machine.SPI0,
			machine.TFT_DC,
			machine.TFT_CS,
			machine.NoPin,
			machine.NoPin,
		)
	*/
)

func main() {

	//machine.SPI0.Configure(machine.SPIConfig{Frequency: 16 * 1e6})

	println("turning on backlight")

	machine.TFT_BACKLIGHT.Configure(machine.PinConfig{machine.PinOutput})

	println("configuring display")
	display.Configure(ili9341.Config{})

	print("width, height == ")
	width, height := display.Size()
	println(width, height)

	black := color.RGBA{0, 0, 0, 255}

	display.FillScreen(black)
	machine.TFT_BACKLIGHT.High()

	white := color.RGBA{255, 255, 255, 255}
	red := color.RGBA{255, 0, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	green := color.RGBA{0, 255, 0, 255}
	display.FillRectangle(0, 0, width/2, height/2, white)
	display.FillRectangle(width/2, 0, width/2, height/2, red)
	display.FillRectangle(0, height/2, width/2, height/2, green)
	display.FillRectangle(width/2, height/2, width/2, height/2, blue)
	display.FillRectangle(width/4, height/4, width/2, height/2, black)
	for {
		time.Sleep(time.Second)
		print(". ")
	}
}
