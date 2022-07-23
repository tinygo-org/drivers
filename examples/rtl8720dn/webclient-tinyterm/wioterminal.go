//go:build wioterminal
// +build wioterminal

package main

import (
	"tinygo.org/x/drivers/examples/ili9341/initdisplay"
	"tinygo.org/x/drivers/ili9341"
)

var (
	display *ili9341.Device
)

func init() {
	display = initdisplay.InitDisplay()
	display.SetRotation(ili9341.Rotation0)
}
