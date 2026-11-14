//go:build arduino_uno_r4_wifi

package main

import (
	"image/color"

	"tinygo.org/x/drivers/unor4matrix"
)

func main() {
	display := unor4matrix.NewOnboard()
	if err := display.Configure(); err != nil {
		println(err.Error())
		return
	}
	width, height := display.Size()
	for y := int16(0); y < height; y++ {
		for x := int16(0); x < width; x++ {
			if (x+y)%2 == 0 {
				display.SetPixel(x, y, color.RGBA{R: 255})
			}
		}
	}
	for {
		if err := display.Display(); err != nil {
			println(err.Error())
			return
		}
	}
}
