package ws2812

import "image/color"

// applyBrightness scales a color by the brightness value.
func applyBrightness(c color.RGBA, brightness uint8) (r, g, b uint8) {
	scale := uint16(brightness) + 1
	r = uint8((uint16(c.R) * scale) >> 8)
	g = uint8((uint16(c.G) * scale) >> 8)
	b = uint8((uint16(c.B) * scale) >> 8)
	return
}
