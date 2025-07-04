package main

// This example shows how to use SSD1306 OLED display driver over I2C and SPI.
//
// Check the `newSSD1306Display()` functions for I2C and SPI initializations.

import (
	"runtime"

	"image/color"
	"time"
)

func main() {

	display := newSSD1306Display()
	display.ClearDisplay()

	w, h := display.Size()
	x := int16(0)
	y := int16(0)
	deltaX := int16(1)
	deltaY := int16(1)

	traceTime := time.Now().UnixMilli() + 1000
	frames := 0
	ms := runtime.MemStats{}

	for {
		pixel := display.GetPixel(x, y)
		c := color.RGBA{255, 255, 255, 255}
		if pixel {
			c = color.RGBA{0, 0, 0, 255}
		}
		display.SetPixel(x, y, c)
		display.Display()

		x += deltaX
		y += deltaY

		if x == 0 || x == w-1 {
			deltaX = -deltaX
		}

		if y == 0 || y == h-1 {
			deltaY = -deltaY
		}

		frames++
		now := time.Now().UnixMilli()
		if now >= traceTime {
			runtime.ReadMemStats(&ms)
			println("TS", now, "| FPS", frames, "| HeapInuse", ms.HeapInuse)
			traceTime = now + 1000
			frames = 0
		}
	}

}
