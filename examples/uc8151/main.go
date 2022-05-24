package main

import (
	"image/color"
	"machine"

	"tinygo.org/x/drivers/uc8151"
)

var display uc8151.Device
var led machine.Pin

func main() {
	led = machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 12000000,
		SCK:       machine.EPD_SCK_PIN,
		SDO:       machine.EPD_SDO_PIN,
	})

	display = uc8151.New(machine.SPI0, machine.EPD_CS_PIN, machine.EPD_DC_PIN, machine.EPD_RESET_PIN, machine.EPD_BUSY_PIN)
	display.Configure(uc8151.Config{
		Rotation: uc8151.ROTATION_270,
		Speed:    uc8151.MEDIUM,
		Blocking: true,
	})
	black := color.RGBA{1, 1, 1, 255}

	display.ClearBuffer()
	display.Display()
	for i := int16(0); i < 37; i++ {
		for j := int16(0); j < 16; j++ {
			if (i+j)%2 == 0 {
				showRect(i*8, j*8, 8, 8, black)
			}
		}
	}

	display.Display()
}

func showRect(x int16, y int16, w int16, h int16, c color.RGBA) {
	for i := x; i < x+w; i++ {
		for j := y; j < y+h; j++ {
			display.SetPixel(i, j, c)
		}
	}
}
