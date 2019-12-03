// Simple driver to create a _screen_ with LED strips
package ledstripdisplay // import "tinygo.org/x/drivers/ledstripdisplay"

import (
	"image/color"

	"tinygo.org/x/drivers"
)

const (
	/**
	 * Order of the LEDstrip in the different layouts
	 *
	 * PARALLEL_1 :
	 *  0  1  2  3
	 *  4  5  6  7
	 *  8  9 10 11
	 * 12 13 14 15
	 *
	 * PARALLEL_2 :
	 * 12 13 14 15
	 *  8  9 10 11
	 *  4  5  6  7
	 *  0  1  2  3
	 *
	 * LAYOUT_Z :
	 *  0  1  2  3
	 *  7  6  5  4
	 *  8  9 10 11
	 * 15 14 13 12
	 *
	 * LAYOUT_S :
	 *  3  2  1  0
	 *  4  5  6  7
	 * 11 10  9  8
	 * 12 13 14 15
	 */
	PARALLEL_1 Layout = 0
	PARALLEL_2 Layout = 1
	LAYOUT_Z   Layout = 2
	LAYOUT_S   Layout = 3

	NO_ROTATION  Rotation = 0
	ROTATION_90  Rotation = 1 // 90 degrees clock-wise rotation
	ROTATION_180 Rotation = 2
	ROTATION_270 Rotation = 3
)

type Layout uint8
type Rotation uint8

// Device holds LEDStriper device and some other information
type Device struct {
	ledstrip drivers.LEDStriper
	width    int16
	height   int16
	layout   Layout
	rotation Rotation
	buffer   []color.RGBA
}

// Config is the configuration for the display
type Config struct {
	Rotation Rotation
}

// New returns a new ledstripdisplay driver given a LEDStriper, layout and rotation
func New(ledstrip drivers.LEDStriper, width, height int16, layout Layout) Device {
	return Device{
		ledstrip: ledstrip,
		width:    width,
		height:   height,
		layout:   layout,
	}
}

// Configure initializes the display with default configuration
func (d *Device) Configure(cfg Config) {
	d.rotation = cfg.Rotation
	d.buffer = make([]color.RGBA, d.width*d.height)
}

// ClearDisplay erases the internal buffer
func (d *Device) ClearDisplay() {
	black := color.RGBA{0, 0, 0, 0}
	for i := int16(0); i < d.width*d.height; i++ {
		d.buffer[i] = black
	}
}

// SetPizel modifies the internal buffer.
func (d *Device) SetPixel(x, y int16, c color.RGBA) {
	w, h := d.Size()
	if x < 0 || y < 0 || x >= w || y >= h {
		return
	}
	if d.rotation == ROTATION_90 {
		x, y = d.width-y-1, x
	} else if d.rotation == ROTATION_180 {
		x = d.width - x - 1
		y = d.height - y - 1
	} else if d.rotation == ROTATION_270 {
		x, y = y, d.height-x-1
	}
	switch d.layout {
	case PARALLEL_1:
		d.buffer[x+y*d.width] = c
		break
	case PARALLEL_2:
		d.buffer[x+(d.height-y-1)*d.width] = c
		break
	case LAYOUT_Z:
		if y%2 == 0 {
			d.buffer[x+y*d.width] = c
		} else {
			d.buffer[(d.width-x-1)+y*d.width] = c
		}
		break
	case LAYOUT_S:
		if y%2 == 0 {
			d.buffer[(d.width-x-1)+y*d.width] = c
		} else {
			d.buffer[x+y*d.width] = c
		}
		break
	}
}

// Display sends the buffer (if any) to the screen.
func (d *Device) Display() error {
	_, err := d.ledstrip.WriteColors(d.buffer)
	return err
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	if d.rotation == NO_ROTATION || d.rotation == ROTATION_180 {
		return d.width, d.height
	}
	return d.height, d.width
}
