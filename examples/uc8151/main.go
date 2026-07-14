package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/uc8151"
)

var display uc8151.Device
var led machine.Pin

func main() {
	led = machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 12 * machine.MHz,
		SCK:       machine.EPD_SCK_PIN,
		SDO:       machine.EPD_SDO_PIN,
	})

	display = uc8151.New(machine.SPI0, machine.EPD_CS_PIN, machine.EPD_DC_PIN, machine.EPD_RESET_PIN, machine.EPD_BUSY_PIN)
	display.Configure(uc8151.Config{
		Rotation:    drivers.Rotation270,
		Speed:       uc8151.TURBO,
		FlickerFree: true,
		Blocking:    false,
	})
	black := color.RGBA{1, 1, 1, 255}

	display.ClearDisplay()

	mod := int16(1)
	for {
		// checkerboard
		for i := int16(0); i < 11; i++ {
			if mod == 1 {
				mod = 0
			} else {
				mod = 1
			}

			display.ClearBuffer()
			for i := int16(0); i < 37; i++ {
				for j := int16(0); j < 16; j++ {
					if (i+j)%2 == mod {
						showRect(i*8, j*8, 8, 8, black)
					}
				}
			}
			display.Display()
			time.Sleep(500 * time.Millisecond)
		}

		// moving line
		for i := int16(16); i < 21; i++ {
			display.ClearBuffer()
			for j := int16(0); j < 16; j++ {
				if (i+j)%2 == 0 {
					showRect(i*8, j*8, 8, 8, black)
				}
				display.Display()
				time.Sleep(250 * time.Millisecond)
			}
		}
	}
}

func showRect(x int16, y int16, w int16, h int16, c color.RGBA) {
	for i := x; i < x+w; i++ {
		for j := y; j < y+h; j++ {
			display.SetPixel(i, j, c)
		}
	}
}
