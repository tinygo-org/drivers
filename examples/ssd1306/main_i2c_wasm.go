//go:build js || wasip1

package main

import (
	"tinygo.org/x/drivers/ssd1306"
)

// A WebAssembly build has no I2C bus. This bus accepts all writes, so the
// example compiles and runs the driver without hardware.
type bus struct{}

func (bus) Tx(addr uint16, w, r []byte) error { return nil }

func newSSD1306Display() *ssd1306.Device {
	display := ssd1306.NewI2C(bus{})
	display.Configure(ssd1306.Config{Width: 128, Height: 64, Address: 0x3C})
	return display
}
