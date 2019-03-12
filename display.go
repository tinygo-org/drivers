package drivers

import "image/color"

type Display interface {
	// Size returns the current size of the display.
	Size() (x, y int16)

	// SetPizel modifies the internal buffer.
	SetPixel(x, y int16, c color.RGBA)

	// Display sends the buffer (if any) to the screen.
	Display() error

	// DisplayPixel sends a single pixel to the screen.
	DisplayPixel(x, y int16, c color.RGBA)

	// DisplayRect sends this part of the buffer to the screen for incremental updates
	// (e.g. button press animations, blinking cursor, etc.).
	DisplayRect(x1, y1, x2, y2 int16)
}
