package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/waveshare-epd/epd2in66b"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

var (
	black = color.RGBA{0, 0, 0, 0xff}
	white = color.RGBA{0xff, 0xff, 0xff, 0xff}
	red   = color.RGBA{0xff, 0, 0, 0xff}
)

func main() {
	machine.Serial.Configure(machine.UARTConfig{})
	time.Sleep(2 * time.Second)

	machine.SPI1.Configure(machine.SPIConfig{
		Frequency: epd2in66b.Baudrate,
	})

	println("started")

	// in case you have a Pico module, you can directly use
	// dev, err := epd2in66b.NewPicoModule()

	display := epd2in66b.New(machine.SPI1)

	cfg := epd2in66b.Config{
		DataPin:       machine.GP8,
		ChipSelectPin: machine.GP9,
		ResetPin:      machine.GP12,
		BusyPin:       machine.GP13,
	}
	err := display.Configure(cfg)
	if err != nil {
		panic(err)
	}

	err = display.Reset()
	if err != nil {
		panic(err)
	}

	println("draw checkerboard")
	drawCheckerBoard(&display)

	println("draw 'hello'")
	tinyfont.WriteLineRotated(&display, &freemono.Bold24pt7b, 40, 10, "Hello!", white, tinyfont.ROTATION_90)
	tinyfont.WriteLineRotated(&display, &freemono.Bold12pt7b, 10, 10, "tinygo rocks", white, tinyfont.ROTATION_90)
	err = display.Display()
	if err != nil {
		panic(err)
	}
}

func drawCheckerBoard(display drivers.Displayer) {
	s := 8
	width, height := display.Size()
	for x := 0; x <= int(width)-s; x += s {
		for y := 0; y <= int(height)-s; y += s {
			c := red
			if (x/s)%2 == (y/s)%2 {
				c = black
			}

			showRect(display, x, y, s, s, c)
		}
	}
}

func showRect(display drivers.Displayer, x int, y int, w int, h int, c color.RGBA) {
	for i := x; i < x+w; i++ {
		for j := y; j < y+h; j++ {
			display.SetPixel(int16(i), int16(j), c)
		}
	}
}
