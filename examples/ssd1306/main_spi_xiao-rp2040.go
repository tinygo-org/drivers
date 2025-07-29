//go:build xiao_rp2040

// This initializes SSD1306 OLED display driver over SPI.
//
// Seeed XIAO RP2040 board + SSD1306 128x64 SPI OLED display.
//
// Wiring:
// - XIAO GND       -> OLED GND
// - XIAO 3v3       -> OLED VCC
// - XIAO D8 (SCK)  -> OLED D0
// - XIAO D10 (SDO) -> OLED D1
// - XIAO D4        -> OLED RES
// - XIAO D5        -> OLED DC
// - XIAO D6        -> OLED CS
//
// For your case:
// - Connect the display to SPI pins on your board.
// - Adjust RES, DC and CS pins as needed.
// - Adjust SPI frequency as needed.
// - Adjust display size as needed.

package main

import (
	"machine"

	"tinygo.org/x/drivers/ssd1306"
)

func newSSD1306Display() *ssd1306.Device {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 50 * machine.MHz,
	})
	display := ssd1306.NewSPI(machine.SPI0, machine.D5, machine.D4, machine.D6)
	display.Configure(ssd1306.Config{
		Width:  128,
		Height: 64,
	})
	return display
}
