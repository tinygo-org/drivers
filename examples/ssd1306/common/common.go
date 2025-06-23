package common

import (
	"runtime"

	"image/color"
	"time"

	"tinygo.org/x/drivers/ssd1306"
)

var ms = runtime.MemStats{}

func Loop(display ssd1306.Device) {
	display.ClearDisplay()
	w, h := display.Size()
	x := int16(0)
	y := int16(0)
	deltaX := int16(1)
	deltaY := int16(1)
	trace := time.Now().UnixMilli() + 1000
	frames := 0
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
		if now >= trace {
			runtime.ReadMemStats(&ms)
			println("TS", now, "| FPS", frames, "| HeapInuse", ms.HeapInuse)
			trace = now + 1000
			frames = 0
		}
	}
}
