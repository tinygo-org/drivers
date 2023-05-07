package neopixel

// Constants for NeoPixels on the Seesaw
// https://github.com/adafruit/Adafruit_Seesaw/blob/master/seesaw_neopixel.h#L49

// PixelType defines the order of primary colors in the NeoPixel data stream, which can vary
// among device types, manufacturers and even different revisions of
// the same item.  The PixelType encodes the per-pixel byte offsets of the red, green
// and blue primaries (plus white, if present) in the data stream.
//
// Below an easier-to-use named version for
// each permutation.  e.g. PixelTypeGRB indicates a NeoPixel-compatible
// device expecting three bytes per pixel, with the first byte
// containing the green value, second containing red and third
// containing blue.  The in-memory representation of a chain of
// NeoPixels is the same as the data-stream order; no re-ordering of
// bytes is required when issuing data to the chain.
//
// Bits 5,4 of this value are the offset (0-3) from the first byte of
// a pixel to the location of the red color byte.  Bits 3,2 are the
// green offset and 1,0 are the blue offset.  If it is an RGBW-type
// device (supporting a white primary in addition to R,G,B), bits 7,6
// are the offset to the white byte...otherwise, bits 7,6 are set to
// the same value as 5,4 (red) to indicate an RGB (not RGBW) device.
//
// i.e. binary representation:
//
//	0bWWRRGGBB for RGBW devices
//	0bRRRRGGBB for RGB
type PixelType uint16

// BlueOffset returns the byte offset for the blue component
func (p PixelType) BlueOffset() int {
	return int(byte(p) & 0b11)
}

// GreenOffset returns the byte offset for the green component
func (p PixelType) GreenOffset() int {
	return int((byte(p) >> 2) & 0b11)
}

// RedOffset returns the byte offset for the red component
func (p PixelType) RedOffset() int {
	return int((byte(p) >> 4) & 0b11)
}

// WhiteOffset returns the byte offset for the white component
func (p PixelType) WhiteOffset() int {
	return int((byte(p) >> 6) & 0b11)
}

// IsRGBW returns true if the pixel has a W component, false otherwise.
func (p PixelType) IsRGBW() bool {
	return p.RedOffset() != p.WhiteOffset()
}

// EncodedLen returns the number of bytes this pixel type needs in encoded form
func (p PixelType) EncodedLen() int {
	if p.IsRGBW() {
		return 4
	}
	return 3
}

// Is800KHz whether the pixel operates at 800 Khz or not. Implicitly runs at 400 Khz otherwise
func (p PixelType) Is800KHz() bool {
	return (uint16(p) & 0xFf00) == SpeedKHz800
}

// PutRGBW encodes color components into a PixelType and appends it to a buffer. Depending on the PixelType it will
// write a different amount of bytes, the count of which will be returned.
func (p PixelType) PutRGBW(buf []byte, color RGBW) int {
	buf[p.RedOffset()] = color.R
	buf[p.GreenOffset()] = color.G
	buf[p.BlueOffset()] = color.B
	if !p.IsRGBW() {
		// if we don't have white, skip it
		return 3
	}
	buf[p.WhiteOffset()] = color.W
	return 4
}

// Rgb NeoPixel permutations; white and red offsets are always same
// Offset:                       W          R          G          B
const (
	PixelTypeRGB PixelType = (0 << 6) | (0 << 4) | (1 << 2) | (2)
	PixelTypeRBG PixelType = (0 << 6) | (0 << 4) | (2 << 2) | (1)
	PixelTypeGRB PixelType = (1 << 6) | (1 << 4) | (0 << 2) | (2)
	PixelTypeGBR PixelType = (2 << 6) | (2 << 4) | (0 << 2) | (1)
	PixelTypeBRG PixelType = (1 << 6) | (1 << 4) | (2 << 2) | (0)
	PixelTypeBGR PixelType = (2 << 6) | (2 << 4) | (1 << 2) | (0)
)

// Rgbw NeoPixel permutations; all 4 offsets are distinct
// Offset:                        W          R          G          B
const (
	PixelTypeWRGB PixelType = (0 << 6) | (1 << 4) | (2 << 2) | (3)
	PixelTypeWRBG PixelType = (0 << 6) | (1 << 4) | (3 << 2) | (2)
	PixelTypeWGRB PixelType = (0 << 6) | (2 << 4) | (1 << 2) | (3)
	PixelTypeWGBR PixelType = (0 << 6) | (3 << 4) | (1 << 2) | (2)
	PixelTypeWBRG PixelType = (0 << 6) | (2 << 4) | (3 << 2) | (1)
	PixelTypeWBGR PixelType = (0 << 6) | (3 << 4) | (2 << 2) | (1)

	PixelTypeRWGB PixelType = (1 << 6) | (0 << 4) | (2 << 2) | (3)
	PixelTypeRWBG PixelType = (1 << 6) | (0 << 4) | (3 << 2) | (2)
	PixelTypeRGWB PixelType = (2 << 6) | (0 << 4) | (1 << 2) | (3)
	PixelTypeRGBW PixelType = (3 << 6) | (0 << 4) | (1 << 2) | (2)
	PixelTypeRBWG PixelType = (2 << 6) | (0 << 4) | (3 << 2) | (1)
	PixelTypeRBGW PixelType = (3 << 6) | (0 << 4) | (2 << 2) | (1)

	PixelTypeGWRB PixelType = (1 << 6) | (2 << 4) | (0 << 2) | (3)
	PixelTypeGWBR PixelType = (1 << 6) | (3 << 4) | (0 << 2) | (2)
	PixelTypeGRWB PixelType = (2 << 6) | (1 << 4) | (0 << 2) | (3)
	PixelTypeGRBW PixelType = (3 << 6) | (1 << 4) | (0 << 2) | (2)
	PixelTypeGBWR PixelType = (2 << 6) | (3 << 4) | (0 << 2) | (1)
	PixelTypeGBRW PixelType = (3 << 6) | (2 << 4) | (0 << 2) | (1)

	PixelTypeBWRG PixelType = (1 << 6) | (2 << 4) | (3 << 2) | (0)
	PixelTypeBWGR PixelType = (1 << 6) | (3 << 4) | (2 << 2) | (0)
	PixelTypeBRWG PixelType = (2 << 6) | (1 << 4) | (3 << 2) | (0)
	PixelTypeBRGW PixelType = (3 << 6) | (1 << 4) | (2 << 2) | (0)
	PixelTypeBGWR PixelType = (2 << 6) | (3 << 4) | (1 << 2) | (0)
	PixelTypeBGRW PixelType = (3 << 6) | (2 << 4) | (1 << 2) | (0)
)

const (
	// SpeedKHz800 represents the encoded value for a 800KHz datastream
	SpeedKHz800 = 0x0000

	// SpeedKHz400 represents the encoded value for a 400KHz datastream
	SpeedKHz400 = 0x0100
)
