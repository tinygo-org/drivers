package main

import (
	"machine"

	"image/color"

	"tinygo.org/x/drivers/st7789"
)

func main() {

	// Example configuration for Adafruit Clue
	// machine.SPI1.Configure(machine.SPIConfig{
	//	Frequency: 8000000,
	//	SCK:       machine.TFT_SCK,
	//	SDO:       machine.TFT_SDO,
	//	SDI:       machine.TFT_SDO,
	//	Mode:      0,
	// })
	// display := st7789.New(machine.SPI1,
	//	machine.TFT_RESET,
	//	machine.TFT_DC,
	//	machine.TFT_CS,
	//	machine.TFT_LITE)

	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 8000000,
		Mode:      0,
	})
	display := st7789.New(machine.SPI0,
		machine.TFT_RESET, // TFT_RESET
		machine.TFT_DC,    // TFT_DC
		machine.TFT_CS,    // TFT_CS
		machine.TFT_LITE)  // TFT_LITE

	display.Configure(st7789.Config{
		Rotation:   st7789.NO_ROTATION,
		RowOffset:  80,
		FrameRate:  st7789.FRAMERATE_111,
		VSyncLines: st7789.MAX_VSYNC_SCANLINES,
	})

	width, height := display.Size()

	white := color.RGBA{255, 255, 255, 255}
	red := color.RGBA{255, 0, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	green := color.RGBA{0, 255, 0, 255}
	black := color.RGBA{0, 0, 0, 255}

	display.FillScreen(black)

	display.FillRectangle(0, 0, width/2, height/2, white)
	display.FillRectangle(width/2, 0, width/2, height/2, red)
	display.FillRectangle(0, height/2, width/2, height/2, green)
	display.FillRectangle(width/2, height/2, width/2, height/2, blue)
	display.FillRectangle(width/4, height/4, width/2, height/2, black)
}
