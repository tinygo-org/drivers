//go:build m5stack_core2

package main

import (
	"image/color"
	"machine"

	axp192 "tinygo.org/x/drivers/axp192/m5stack-core2-axp192"
	"tinygo.org/x/drivers/ft6336"
	"tinygo.org/x/drivers/i2csoft"
	"tinygo.org/x/drivers/ili9341"
	"tinygo.org/x/drivers/touch"
)

// InitDisplay initializes the display of each board.
func initDevices() (touchPaintDisplay, touch.Pointer, error) {
	machine.SPI2.Configure(machine.SPIConfig{
		SCK:       machine.LCD_SCK_PIN,
		SDO:       machine.LCD_SDO_PIN,
		SDI:       machine.LCD_SDI_PIN,
		Frequency: 40e6,
	})

	i2c := i2csoft.New(machine.SCL0_PIN, machine.SDA0_PIN)
	i2c.Configure(i2csoft.I2CConfig{Frequency: 100e3})

	axp := axp192.New(i2c)
	led := axp.LED
	led.Low()

	display := ili9341.NewSPI(
		machine.SPI2,
		machine.LCD_DC_PIN,
		machine.LCD_SS_PIN,
		machine.NoPin,
	)

	// configure display
	display.Configure(ili9341.Config{
		Width:            320,
		Height:           240,
		DisplayInversion: true,
	})
	display.FillScreen(color.RGBA{255, 255, 255, 255})

	display.SetRotation(ili9341.Rotation0Mirror)

	resistiveTouch := ft6336.New(i2c, machine.Pin(39))
	resistiveTouch.Configure(ft6336.Config{})
	resistiveTouch.SetPeriodActive(0x00)

	return display, resistiveTouch, nil
}
