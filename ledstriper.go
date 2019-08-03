package drivers

import "image/color"

type LedStriper interface {
	// Write the raw bitstring out using the specific led strip protocol.
	Write(buf []byte) (n int, err error)

	// Write the given color slice out using the specific led strip protocol and format.
	WriteColors(buf []color.RGBA) (n int, err error)
}
