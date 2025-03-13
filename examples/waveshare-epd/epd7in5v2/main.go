package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/waveshare-epd/epd7in5v2"
)

// example for writing to a e-ink display
func main() {
	// Pins are for an Adafruit Feather nRF52840 Express
	bus := machine.SPI0
	cs := machine.D5
	dc := machine.D6
	rst := machine.D9
	busy := machine.D10
	pwr := machine.D11

	panel := epd7in5v2.New(cs, rst, dc, busy, pwr, bus)
	err := panel.Configure(machine.SPI0_SCK_PIN, machine.SPI0_SDO_PIN)
	if err != nil {
		panic(err)
	}

	panel.Reset()

	err = panel.Init()
	if err != nil {
		panic(err)
	}

	buf := epd7in5v2.NewImageBuffer()

	// Draw some stripes
	var color byte
	for y := 0; y < epd7in5v2.HEIGHT; y++ {
		if y/60%2 == 0 {
			color = 0xFF
		} else {
			color = 0x00
		}
		for x := 0; x < epd7in5v2.BYTE_WIDTH; x++ {
			buf[y*epd7in5v2.BYTE_WIDTH+x] = color
		}
	}

	err = panel.Draw(buf)
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	err = panel.Clear()
	if err != nil {
		panic(err)
	}

	err = panel.DeepSleep()
	if err != nil {
		panic(err)
	}
}
