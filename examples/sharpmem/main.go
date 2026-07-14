package main

import (
	"image/color"
	"machine"
	"math/rand/v2"
	"time"

	"tinygo.org/x/drivers/sharpmem"
)

var (
	// example wiring using a nice!view and nice!nano:
	// (view)   (nano)
	// MOSI --> P0.24
	// SCK ---> P0.22
	// GND ---> GND
	// VCC ---> 3.3V
	// CS ----> P0.06

	spi = machine.SPI0

	sckPin = machine.SPI0_SCK_PIN // SCK
	sdoPin = machine.SPI0_SDO_PIN // MOSI
	sdiPin = machine.SPI0_SDI_PIN // (any pin)

	csPin = machine.P0_06 // CS
)

func main() {
	time.Sleep(time.Second)

	err := spi.Configure(machine.SPIConfig{
		Frequency: 2000000,
		SCK:       sckPin,
		SDO:       sdoPin,
		SDI:       sdiPin,
		Mode:      0,
		LSBFirst:  true,
	})
	if err != nil {
		println("spi.Configure() failed, error:", err.Error())
		return
	}

	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	display := sharpmem.New(spi, csPin)

	cfg := sharpmem.ConfigLS011B7DH03
	display.Configure(cfg)

	// clear the display before first use
	err = display.Clear()
	if err != nil {
		println("display.Clear() failed, error:", err.Error())
		return
	}

	// random boxes pop into and out of existence
	for {
		x0 := int16(rand.IntN(int(cfg.Width - 7)))
		y0 := int16(rand.IntN(int(cfg.Height - 7)))

		for x2 := int16(0); x2 < 16; x2++ {
			x2 := x2
			c := color.RGBA{R: 255, G: 255, B: 255, A: 255}

			if x2 >= 8 {
				// effectively erases the box after it showed up
				x2 = x2 - 8
				c = color.RGBA{R: 0, G: 0, B: 0, A: 255}
			}

			for x := int16(0); x < x2; x++ {
				for y := int16(0); y < 8; y++ {
					display.SetPixel(x0+x, y0+y, c)
				}
			}

			err = display.Display()
			if err != nil {
				println("display.Display() failed, error:", err.Error())
				continue
			}

			time.Sleep(33 * time.Millisecond)
		}
	}
}
