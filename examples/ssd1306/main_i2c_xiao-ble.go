//go:build xiao_ble

// This initializes SSD1306 OLED display driver over I2C.
//
// Seeed XIAO BLE board + SSD1306 128x32 I2C OLED display.
//
// Wiring:
// - XIAO GND      -> OLED GND
// - XIAO 3v3      -> OLED VCC
// - XIAO D4 (SDA) -> OLED SDA
// - XIAO D5 (SCL) -> OLED SCK
//
// For your case:
// - Connect the display to I2C pins on your board.
// - Adjust I2C address and display size as needed.

package main

import (
	"machine"

	"tinygo.org/x/drivers/ssd1306"
)

func newSSD1306Display() *ssd1306.Device {
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: 400 * machine.KHz,
		SDA:       machine.SDA0_PIN,
		SCL:       machine.SCL0_PIN,
	})
	display := ssd1306.NewI2C(machine.I2C0)
	display.Configure(ssd1306.Config{
		Address: ssd1306.Address_128_32, // or ssd1306.Address
		Width:   128,
		Height:  32, // or 64
	})
	return display
}
