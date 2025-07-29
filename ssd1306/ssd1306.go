// Package ssd1306 implements a driver for the SSD1306 led matrix controller, it comes in various colors and screen sizes.
//
// Datasheet: https://cdn-shop.adafruit.com/datasheets/SSD1306.pdf
package ssd1306 // import "tinygo.org/x/drivers/ssd1306"

import (
	"errors"
	"image/color"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/pixel"
)

var (
	errBufferSize = errors.New("invalid size buffer")
	errOutOfRange = errors.New("out of screen range")
)

type ResetValue [2]byte

// Device wraps I2C or SPI connection.
type Device struct {
	bus       Buser
	buffer    []byte
	width     int16
	height    int16
	vccState  VccMode
	canReset  bool
	resetCol  ResetValue
	resetPage ResetValue
	rotation  drivers.Rotation
}

// Config is the configuration for the display
type Config struct {
	Width    int16
	Height   int16
	VccState VccMode
	Address  uint16
	// ResetCol and ResetPage are used to reset the screen to 0x0
	// This is useful for some screens that have a different size than 128x64
	// For example, the Thumby's screen is 72x40
	// The default values are normally set automatically based on the size.
	// If you're using a different size, you might need to set these values manually.
	ResetCol  ResetValue
	ResetPage ResetValue
	Rotation  drivers.Rotation
}

type Buser interface {
	configure(address uint16, size int16) []byte // configure the bus and return the image buffer to use
	command(cmd uint8) error                     // send a command to the display
	flush() error                                // send the image to the display, faster than "tx()" in i2c case since avoids slice copy
	tx(data []byte, isCommand bool) error        // generic transmit function
}

type VccMode uint8

// Configure initializes the display with default configuration
func (d *Device) Configure(cfg Config) {
	var zeroReset ResetValue
	if cfg.Width != 0 {
		d.width = cfg.Width
	} else {
		d.width = 128
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	} else {
		d.height = 64
	}
	if cfg.VccState != 0 {
		d.vccState = cfg.VccState
	} else {
		d.vccState = SWITCHCAPVCC
	}
	if cfg.ResetCol != zeroReset {
		d.resetCol = cfg.ResetCol
	} else {
		d.resetCol = ResetValue{0, uint8(d.width - 1)}
	}
	if cfg.ResetPage != zeroReset {
		d.resetPage = cfg.ResetPage
	} else {
		d.resetPage = ResetValue{0, uint8(d.height/8) - 1}
	}
	d.canReset = cfg.Address != 0 || d.width != 128 || d.height != 64 // I2C or not 128x64

	d.buffer = d.bus.configure(cfg.Address, d.width*d.height/8)

	time.Sleep(100 * time.Nanosecond)
	d.Command(DISPLAYOFF)
	d.Command(SETDISPLAYCLOCKDIV)
	d.Command(0x80)
	d.Command(SETMULTIPLEX)
	d.Command(uint8(d.height - 1))
	d.Command(SETDISPLAYOFFSET)
	d.Command(0x0)
	d.Command(SETSTARTLINE | 0x0)
	d.Command(CHARGEPUMP)
	if d.vccState == EXTERNALVCC {
		d.Command(0x10)
	} else {
		d.Command(0x14)
	}
	d.Command(MEMORYMODE)
	d.Command(0x00)

	d.SetRotation(cfg.Rotation)

	if (d.width == 128 && d.height == 64) || (d.width == 64 && d.height == 48) { // 128x64 or 64x48
		d.Command(SETCOMPINS)
		d.Command(0x12)
		d.Command(SETCONTRAST)
		if d.vccState == EXTERNALVCC {
			d.Command(0x9F)
		} else {
			d.Command(0xCF)
		}
	} else if d.width == 128 && d.height == 32 { // 128x32
		d.Command(SETCOMPINS)
		d.Command(0x02)
		d.Command(SETCONTRAST)
		d.Command(0x8F)
	} else if d.width == 96 && d.height == 16 { // 96x16
		d.Command(SETCOMPINS)
		d.Command(0x2)
		d.Command(SETCONTRAST)
		if d.vccState == EXTERNALVCC {
			d.Command(0x10)
		} else {
			d.Command(0xAF)
		}
	} else {
		// fail silently, it might work
		println("there's no configuration for this display's size")
	}

	d.Command(SETPRECHARGE)
	if d.vccState == EXTERNALVCC {
		d.Command(0x22)
	} else {
		d.Command(0xF1)
	}
	d.Command(SETVCOMDETECT)
	d.Command(0x40)
	d.Command(DISPLAYALLON_RESUME)
	d.Command(NORMALDISPLAY)
	d.Command(DEACTIVATE_SCROLL)
	d.Command(DISPLAYON)

}

// Command sends a command to the display
func (d *Device) Command(command uint8) {
	d.bus.command(command)
}

// Tx sends data to the display; if isCommand is false, this also updates the image buffer.
func (d *Device) Tx(data []byte, isCommand bool) error {
	return d.bus.tx(data, isCommand)
}

// ClearBuffer clears the image buffer
func (d *Device) ClearBuffer() {
	for i := 0; i < len(d.buffer); i++ {
		d.buffer[i] = 0
	}
}

// ClearDisplay clears the image buffer and clear the display
func (d *Device) ClearDisplay() {
	d.ClearBuffer()
	d.Display()
}

// Display sends the whole buffer to the screen
func (d *Device) Display() error {
	// Reset the screen to 0x0
	// This works fine with I2C
	// In the 128x64 (SPI) screen resetting to 0x0 after 128 times corrupt the buffer
	// Since we're printing the whole buffer, avoid resetting it in this case
	if d.canReset {
		d.Command(COLUMNADDR)
		d.Command(d.resetCol[0])
		d.Command(d.resetCol[1])
		d.Command(PAGEADDR)
		d.Command(d.resetPage[0])
		d.Command(d.resetPage[1])
	}

	return d.bus.flush()
}

// SetPixel enables or disables a pixel in the buffer
// color.RGBA{0, 0, 0, 255} is consider transparent, anything else
// with enable a pixel on the screen
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return
	}
	byteIndex := x + (y/8)*d.width
	if c.R != 0 || c.G != 0 || c.B != 0 {
		d.buffer[byteIndex] |= 1 << uint8(y%8)
	} else {
		d.buffer[byteIndex] &^= 1 << uint8(y%8)
	}
}

// GetPixel returns if the specified pixel is on (true) or off (false)
func (d *Device) GetPixel(x int16, y int16) bool {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return false
	}
	byteIndex := x + (y/8)*d.width
	return (d.buffer[byteIndex] >> uint8(y%8) & 0x1) == 1
}

// SetBuffer changes the whole buffer at once
func (d *Device) SetBuffer(buffer []byte) error {
	if len(buffer) != len(d.buffer) {
		return errBufferSize
	}
	copy(d.buffer, buffer)
	return nil
}

// GetBuffer returns the whole buffer
func (d *Device) GetBuffer() []byte {
	return d.buffer
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return d.width, d.height
}

// DrawBitmap copies the bitmap to the screen at the given coordinates.
func (d *Device) DrawBitmap(x, y int16, bitmap pixel.Image[pixel.Monochrome]) error {
	width, height := bitmap.Size()
	if x < 0 || x+int16(width) > d.width || y < 0 || y+int16(height) > d.height {
		return errOutOfRange
	}

	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			d.SetPixel(x+int16(i), y+int16(j), bitmap.Get(i, j).RGBA())
		}
	}

	return nil
}

// Rotation returns the currently configured rotation.
func (d *Device) Rotation() drivers.Rotation {
	return d.rotation
}

// SetRotation changes the rotation of the device (clock-wise).
func (d *Device) SetRotation(rotation drivers.Rotation) error {
	d.rotation = rotation
	switch d.rotation {
	case drivers.Rotation0:
		d.Command(SEGREMAP | 0x1) // Reverse horizontal mapping
		d.Command(COMSCANDEC)     // Reverse vertical mapping
	case drivers.Rotation180:
		d.Command(SEGREMAP)   // Normal horizontal mapping
		d.Command(COMSCANINC) // Normal vertical mapping
	// nothing to do
	default:
		d.Command(SEGREMAP | 0x1) // Reverse horizontal mapping
		d.Command(COMSCANDEC)     // Reverse vertical mapping
	}
	return nil
}

// Set the sleep mode for this display. When sleeping, the panel uses a lot
// less power. The display won't show an image anymore, but the memory contents
// should be kept.
func (d *Device) Sleep(sleepEnabled bool) error {
	if sleepEnabled {
		d.Command(DISPLAYOFF)
	} else {
		d.Command(DISPLAYON)
	}
	return nil
}

// FillRectangle fills a rectangle at a given coordinates with a color
func (d *Device) FillRectangle(x, y, width, height int16, c color.RGBA) error {
	dw, dh := d.Size()

	if x < 0 || y < 0 || width <= 0 || height <= 0 ||
		x >= d.width || (x+width) > dw || y >= dh || (y+height) > dh {
		return errOutOfRange
	}

	if x+width == dw && y+height == dh && c.R == 0 && c.G == 0 && c.B == 0 {
		d.ClearDisplay()
		return nil
	}

	for i := x; i < x+width; i++ {
		for j := y; j < y+height; j++ {
			d.SetPixel(i, j, c)
		}
	}

	return nil
}

// SetScroll sets the vertical scrolling for the display, which is a NOP for this display.
func (d *Device) SetScroll(line int16) {
	return
}
