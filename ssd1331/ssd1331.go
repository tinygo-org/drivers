// Package ssd1331 implements a driver for the SSD1331 TFT color displays.
//
// Datasheet: https://www.crystalfontz.com/controllers/SolomonSystech/SSD1331/381/
//
package ssd1331 // import "tinygo.org/x/drivers/ssd1331"

import (
	"image/color"
	"machine"

	"errors"
	"time"

	"tinygo.org/x/drivers"
)

type Model uint8
type Rotation uint8

// Device wraps an SPI connection.
type Device struct {
	bus         drivers.SPI
	dcPin       machine.Pin
	resetPin    machine.Pin
	csPin       machine.Pin
	width       int16
	height      int16
	batchLength int16
	isBGR       bool
	batchData   []uint8
}

// Config is the configuration for the display
type Config struct {
	Width  int16
	Height int16
}

// New creates a new SSD1331 connection. The SPI wire must already be configured.
func New(bus drivers.SPI, resetPin, dcPin, csPin machine.Pin) Device {
	dcPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	resetPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return Device{
		bus:      bus,
		dcPin:    dcPin,
		resetPin: resetPin,
		csPin:    csPin,
	}
}

// Configure initializes the display with default configuration
func (d *Device) Configure(cfg Config) {
	if cfg.Width != 0 {
		d.width = cfg.Width
	} else {
		d.width = 96
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	} else {
		d.height = 64
	}

	d.batchLength = d.width
	if d.height > d.width {
		d.batchLength = d.height
	}
	d.batchLength += d.batchLength & 1
	d.batchData = make([]uint8, d.batchLength*2)

	// reset the device
	d.resetPin.High()
	time.Sleep(100 * time.Millisecond)
	d.resetPin.Low()
	time.Sleep(100 * time.Millisecond)
	d.resetPin.High()
	time.Sleep(200 * time.Millisecond)

	// Initialization
	d.Command(DISPLAYOFF)
	d.Command(SETREMAP)
	d.Command(0x72) // RGB
	//d.Command(0x76) // BGR
	d.Command(STARTLINE)
	d.Command(0x0)
	d.Command(DISPLAYOFFSET)
	d.Command(0x0)
	d.Command(NORMALDISPLAY)
	d.Command(SETMULTIPLEX)
	d.Command(0x3F)
	d.Command(SETMASTER)
	d.Command(0x8E)
	d.Command(POWERMODE)
	d.Command(0x0B)
	d.Command(PRECHARGE)
	d.Command(0x31)
	d.Command(CLOCKDIV)
	d.Command(0xF0)
	d.Command(PRECHARGEA)
	d.Command(0x64)
	d.Command(PRECHARGEB)
	d.Command(0x78)
	d.Command(PRECHARGEC)
	d.Command(0x64)
	d.Command(PRECHARGELEVEL)
	d.Command(0x3A)
	d.Command(VCOMH)
	d.Command(0x3E)
	d.Command(MASTERCURRENT)
	d.Command(0x06)
	d.Command(CONTRASTA)
	d.Command(0x91)
	d.Command(CONTRASTB)
	d.Command(0x50)
	d.Command(CONTRASTC)
	d.Command(0x7D)
	d.Command(DISPLAYON)
}

// Display does nothing, there's no buffer as it might be too big for some boards
func (d *Device) Display() error {
	return nil
}

// SetPixel sets a pixel in the screen
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || y < 0 || x >= d.width || y >= d.height {
		return
	}
	d.FillRectangle(x, y, 1, 1, c)
}

// setWindow prepares the screen to be modified at a given rectangle
func (d *Device) setWindow(x, y, w, h int16) {
	/*d.Tx([]uint8{SETCOLUMN}, true)
	d.Tx([]uint8{uint8(x), uint8(x + w - 1)}, false)
	d.Tx([]uint8{SETROW}, true)
	d.Tx([]uint8{uint8(y), uint8(y + h - 1)}, false)*/
	d.Command(SETCOLUMN)
	d.Command(uint8(x))
	d.Command(uint8(x + w - 1))
	d.Command(SETROW)
	d.Command(uint8(y))
	d.Command(uint8(y + h - 1))
}

// FillRectangle fills a rectangle at a given coordinates with a color
func (d *Device) FillRectangle(x, y, width, height int16, c color.RGBA) error {
	if x < 0 || y < 0 || width <= 0 || height <= 0 ||
		x >= d.width || (x+width) > d.width || y >= d.height || (y+height) > d.height {
		return errors.New("rectangle coordinates outside display area")
	}
	d.setWindow(x, y, width, height)
	c565 := RGBATo565(c)
	c1 := uint8(c565 >> 8)
	c2 := uint8(c565)

	var i int16
	for i = 0; i < d.batchLength; i++ {
		d.batchData[i*2] = c1
		d.batchData[i*2+1] = c2
	}
	i = width * height
	for i > 0 {
		if i >= d.batchLength {
			d.Tx(d.batchData, false)
		} else {
			d.Tx(d.batchData[:i*2], false)
		}
		i -= d.batchLength
	}
	return nil
}

// FillRectangle fills a rectangle at a given coordinates with a buffer
func (d *Device) FillRectangleWithBuffer(x, y, width, height int16, buffer []color.RGBA) error {
	if x < 0 || y < 0 || width <= 0 || height <= 0 ||
		x >= d.width || (x+width) > d.width || y >= d.height || (y+height) > d.height {
		return errors.New("rectangle coordinates outside display area")
	}
	k := width * height
	l := int16(len(buffer))
	if k != l {
		return errors.New("buffer length does not match with rectangle size")
	}

	d.setWindow(x, y, width, height)

	offset := int16(0)
	for k > 0 {
		for i := int16(0); i < d.batchLength; i++ {
			if offset+i < l {
				c565 := RGBATo565(buffer[offset+i])
				c1 := uint8(c565 >> 8)
				c2 := uint8(c565)
				d.batchData[i*2] = c1
				d.batchData[i*2+1] = c2
			}
		}
		if k >= d.batchLength {
			d.Tx(d.batchData, false)
		} else {
			d.Tx(d.batchData[:k*2], false)
		}
		k -= d.batchLength
		offset += d.batchLength
	}
	return nil
}

// DrawFastVLine draws a vertical line faster than using SetPixel
func (d *Device) DrawFastVLine(x, y0, y1 int16, c color.RGBA) {
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	d.FillRectangle(x, y0, 1, y1-y0+1, c)
}

// DrawFastHLine draws a horizontal line faster than using SetPixel
func (d *Device) DrawFastHLine(x0, x1, y int16, c color.RGBA) {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	d.FillRectangle(x0, y, x1-x0+1, 1, c)
}

// FillScreen fills the screen with a given color
func (d *Device) FillScreen(c color.RGBA) {
	d.FillRectangle(0, 0, d.width, d.height, c)
}

// SetContrast sets the three contrast values (A, B & C)
func (d *Device) SetContrast(contrastA, contrastB, contrastC uint8) {
	d.Command(CONTRASTA)
	d.Command(contrastA)
	d.Command(CONTRASTB)
	d.Command(contrastB)
	d.Command(CONTRASTC)
	d.Command(contrastC)
}

// Command sends a command to the display
func (d *Device) Command(command uint8) {
	d.Tx([]byte{command}, true)
}

// Command sends a data to the display
func (d *Device) Data(data uint8) {
	d.Tx([]byte{data}, false)
}

// Tx sends data to the display
func (d *Device) Tx(data []byte, isCommand bool) {
	d.dcPin.Set(!isCommand)
	d.bus.Tx(data, nil)
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return d.width, d.height
}

// IsBGR changes the color mode (RGB/BGR)
func (d *Device) IsBGR(bgr bool) {
	d.isBGR = bgr
}

// RGBATo565 converts a color.RGBA to uint16 used in the display
func RGBATo565(c color.RGBA) uint16 {
	r, g, b, _ := c.RGBA()
	return uint16((r & 0xF800) +
		((g & 0xFC00) >> 5) +
		((b & 0xF800) >> 11))
}
