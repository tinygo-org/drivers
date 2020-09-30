package drivers

import (
	"image/color"
)

// LEDArray is an array of RGB LEDs. It may have any shape, but in general it is
// a strip of daisy-chained LEDs.
type LEDArray interface {
	// WriteColors updates all LEDs in the LED strip to the given RGB color. It
	// depends on the protocol what happens when you do not provide a
	// correctly-sized slice of colors.
	WriteColors(buf []color.RGBA) error
}
