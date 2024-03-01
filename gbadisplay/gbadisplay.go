// Package gbadisplay implements a simple driver for the GameBoy Advance
// display.
package gbadisplay

import (
	"device/gba"
	"errors"
	"image/color"
	"runtime/volatile"
	"unsafe"

	"tinygo.org/x/drivers/pixel"
)

// Image buffer type used by the GameBoy Advance.
type Image = pixel.Image[pixel.RGB555]

const (
	displayWidth  = 240
	displayHeight = 160
)

var (
	errOutOfBounds = errors.New("rectangle coordinates outside display area")
)

type Device struct{}

// New returns a new GameBoy Advance display object.
func New() Device {
	return Device{}
}

var displayFrameBuffer = (*[160 * 240]volatile.Register16)(unsafe.Pointer(uintptr(gba.MEM_VRAM)))

type Config struct {
	// TODO: add more display modes here.
}

// Configure the display as a regular 15bpp framebuffer.
func (d Device) Configure(config Config) {
	// Use video mode 3 (in BG2, a 16bpp bitmap in VRAM) and Enable BG2.
	gba.DISP.DISPCNT.Set(gba.DISPCNT_BGMODE_3<<gba.DISPCNT_BGMODE_Pos |
		gba.DISPCNT_SCREENDISPLAY_BG2_ENABLE<<gba.DISPCNT_SCREENDISPLAY_BG2_Pos)
}

// Size returns the fixed size of this display.
func (d Device) Size() (x, y int16) {
	return displayWidth, displayHeight
}

// Display is a no-op: the display framebuffer is modified directly.
func (d Device) Display() error {
	// Nothing to do here.
	return nil
}

// SetPixel changes the pixel at (x, y) to the given color.
func (d Device) SetPixel(x, y int16, c color.RGBA) {
	if x < 0 || y < 0 || x >= displayWidth || y > displayHeight {
		// Out of bounds, so ignore.
		return
	}
	val := pixel.NewColor[pixel.RGB555](c.R, c.G, c.B)
	displayFrameBuffer[(int(y))*240+int(x)].Set(uint16(val))
}

// DrawBitmap updates the rectangle at (x, y) to the image stored in buf.
func (d Device) DrawBitmap(x, y int16, buf Image) error {
	width, height := buf.Size()
	if x < 0 || y < 0 || int(x)+width > displayWidth || int(y)+height > displayHeight {
		return errOutOfBounds
	}

	// TODO: try to do a 4-byte memcpy if possible. That should significantly
	// speed up the copying of this image.
	for bufY := 0; bufY < int(height); bufY++ {
		for bufX := 0; bufX < int(width); bufX++ {
			val := buf.Get(bufX, bufY)
			displayFrameBuffer[(int(y)+bufY)*240+int(x)+bufX].Set(uint16(val))
		}
	}

	return nil
}
