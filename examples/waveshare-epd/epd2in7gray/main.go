// Four-shade demo for the 2.7in V1 HAT on a Waveshare RP2040-PiZero.
// Flash with tinygo flash -target=waveshare-rp2040-pizero ./examples/waveshare-epd/epd2in7gray
package main

import (
	"image/color"
	"machine"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/waveshare-epd/epd2in7gray"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

func traceStage(stage string) {
	println("epd2in7gray:", stage)
}

func main() {
	if err := run(); err != nil {
		println("epd2in7gray:", err.Error())
	}
}

func run() error {
	if err := machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 4_000_000,
		SCK:       machine.GPIO10,
		SDO:       machine.GPIO11,
		SDI:       machine.GPIO12,
		Mode:      0,
	}); err != nil {
		return err
	}
	traceStage("SPI1 configured")
	cs, dc, rst, busy := machine.GPIO8, machine.GPIO25, machine.GPIO17, machine.GPIO24
	for _, pin := range []machine.Pin{cs, dc, rst} {
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
	busy.Configure(machine.PinConfig{Mode: machine.PinInput})
	traceStage("GPIO pins configured")
	display := epd2in7gray.New(machine.SPI1, cs, dc, rst, busy)
	if err := display.Configure(epd2in7gray.Config{
		Rotation:    drivers.Rotation270,
		OnInitStage: traceStage,
	}); err != nil {
		return err
	}
	if err := display.ClearDisplay(); err != nil {
		return err
	}
	traceStage("white refresh complete")
	drawSwatches(display)
	traceStage("swatches buffered")
	if err := display.Display(); err != nil {
		return err
	}
	traceStage("four-shade refresh complete")
	println("left to right: black, dark gray, light gray, white")
	if err := display.DeepSleep(); err != nil {
		return err
	}
	traceStage("deep sleep complete")
	return nil
}

func drawSwatches(display *epd2in7gray.Device) {
	black := color.RGBA{A: 255}
	w, h := display.Size()
	tinyfont.WriteLine(display, &freemono.Regular9pt7b, 4, 18, "TOP LEFT", black)
	for i, label := range []string{"0", "1", "2", "3"} {
		left := int16(i) * w / 4
		right := int16(i+1) * w / 4
		c := uint8(i * 85)
		shade := color.RGBA{c, c, c, 255}
		for y := int16(30); y < h-30; y++ {
			for x := left + 2; x < right-2; x++ {
				border := x == left+2 || x == right-3 || y == 30 || y == h-31
				pixel := shade
				if border {
					pixel = black
				}
				display.SetPixel(x, y, pixel)
			}
		}
		tinyfont.WriteLine(display, &freemono.Regular9pt7b, left+24, h-10, label, black)
	}
}
