//go:build wioterminal
// +build wioterminal

package main

import (
	"machine"

	"tinygo.org/x/drivers/examples/ili9341/initdisplay"
	"tinygo.org/x/drivers/ili9341"
)

var (
	display *ili9341.Device
)

func init() {
	display = initdisplay.InitDisplay()
	display.SetRotation(ili9341.Rotation270)

	// override
	led = machine.LCD_BACKLIGHT
}
