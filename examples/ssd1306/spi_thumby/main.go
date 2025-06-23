// This example using the SSD1306 OLED display over SPI on the Thumby board
// A very tiny 72x40 display.
package main

import (
	"machine"

	"tinygo.org/x/drivers/examples/ssd1306/common"
	"tinygo.org/x/drivers/ssd1306"
)

func main() {
	machine.SPI0.Configure(machine.SPIConfig{})
	display := ssd1306.NewSPI(machine.SPI0, machine.THUMBY_DC_PIN, machine.THUMBY_RESET_PIN, machine.THUMBY_CS_PIN)
	display.Configure(ssd1306.Config{
		Width:     72,
		Height:    40,
		ResetCol:  ssd1306.ResetValue{28, 99},
		ResetPage: ssd1306.ResetValue{0, 5},
	})

	common.Loop(display)
}
