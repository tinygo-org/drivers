// Package st7735 implements a driver for the ST7735 TFT displays, it comes in various screen sizes.
//
// Datasheet: https://www.crystalfontz.com/controllers/Sitronix/ST7735R/319/
package st7735 // import "tinygo.org/x/drivers/st7735"

import (
	"image/color"
	"machine"
	"time"

	"errors"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/pixel"
)

type Model uint8

// Rotation controls the rotation used by the display.
//
// Deprecated: use drivers.Rotation instead.
type Rotation = drivers.Rotation

// Pixel formats supported by the st7735 driver.
type Color interface {
	pixel.RGB444BE | pixel.RGB565BE

	pixel.BaseColor
}

var (
	errOutOfBounds = errors.New("rectangle coordinates outside display area")
)

// Device wraps an SPI connection.
type Device = DeviceOf[pixel.RGB565BE]

// DeviceOf is a generic version of Device, which supports different pixel
// formats.
type DeviceOf[T Color] struct {
	bus          drivers.SPI
	dcPin        machine.Pin
	resetPin     machine.Pin
	csPin        machine.Pin
	blPin        machine.Pin
	width        int16
	height       int16
	columnOffset int16
	rowOffset    int16
	rotation     drivers.Rotation
	batchLength  int16
	model        Model
	isBGR        bool
	batchData    pixel.Image[T] // "image" with width, height of (batchLength, 1)
}

// Config is the configuration for the display
type Config struct {
	Width        int16
	Height       int16
	Rotation     drivers.Rotation
	Model        Model
	RowOffset    int16
	ColumnOffset int16
}

// New creates a new ST7735 connection. The SPI wire must already be configured.
func New(bus drivers.SPI, resetPin, dcPin, csPin, blPin machine.Pin) Device {
	return NewOf[pixel.RGB565BE](bus, resetPin, dcPin, csPin, blPin)
}

// NewOf creates a new ST7735 connection with a particular pixel format. The SPI
// wire must already be configured.
func NewOf[T Color](bus drivers.SPI, resetPin, dcPin, csPin, blPin machine.Pin) DeviceOf[T] {
	dcPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	resetPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	blPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return DeviceOf[T]{
		bus:      bus,
		dcPin:    dcPin,
		resetPin: resetPin,
		csPin:    csPin,
		blPin:    blPin,
	}
}

// Configure initializes the display with default configuration
func (d *DeviceOf[T]) Configure(cfg Config) {
	d.model = cfg.Model
	if cfg.Width != 0 {
		d.width = cfg.Width
	} else {
		if d.model == MINI80x160 {
			d.width = 80
		} else {
			d.width = 128
		}
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	} else {
		d.height = 160
	}
	d.rotation = cfg.Rotation
	d.rowOffset = cfg.RowOffset
	d.columnOffset = cfg.ColumnOffset

	d.batchLength = d.width
	if d.height > d.width {
		d.batchLength = d.height
	}
	d.batchLength += d.batchLength & 1
	d.batchData = pixel.NewImage[T](int(d.batchLength), 1)

	// reset the device
	d.resetPin.High()
	time.Sleep(5 * time.Millisecond)
	d.resetPin.Low()
	time.Sleep(20 * time.Millisecond)
	d.resetPin.High()
	time.Sleep(150 * time.Millisecond)

	// Common initialization
	d.Command(SWRESET)
	time.Sleep(150 * time.Millisecond)
	d.Command(SLPOUT)
	time.Sleep(500 * time.Millisecond)
	d.Command(FRMCTR1)
	d.Data(0x01)
	d.Data(0x2C)
	d.Data(0x2D)
	d.Command(FRMCTR2)
	d.Data(0x01)
	d.Data(0x2C)
	d.Data(0x2D)
	d.Command(FRMCTR3)
	d.Data(0x01)
	d.Data(0x2C)
	d.Data(0x2D)
	d.Data(0x01)
	d.Data(0x2C)
	d.Data(0x2D)
	d.Command(INVCTR)
	d.Data(0x07)
	d.Command(PWCTR1)
	d.Data(0xA2)
	d.Data(0x02)
	d.Data(0x84)
	d.Command(PWCTR2)
	d.Data(0xC5)
	d.Command(PWCTR3)
	d.Data(0x0A)
	d.Data(0x00)
	d.Command(PWCTR4)
	d.Data(0x8A)
	d.Data(0x2A)
	d.Command(PWCTR5)
	d.Data(0x8A)
	d.Data(0xEE)
	d.Command(VMCTR1)
	d.Data(0x0E)

	// Set the color format depending on the generic type.
	d.Command(COLMOD)
	var zeroColor T
	switch any(zeroColor).(type) {
	case pixel.RGB444BE:
		d.Data(0x03) // 12 bits per pixel
	default:
		d.Data(0x05) // 16 bits per pixel
	}

	if d.model == GREENTAB {
		d.InvertColors(false)
	} else if d.model == MINI80x160 {
		d.isBGR = true
		d.InvertColors(true)
	}

	// common color adjustment
	d.Command(GMCTRP1)
	d.Data(0x02)
	d.Data(0x1C)
	d.Data(0x07)
	d.Data(0x12)
	d.Data(0x37)
	d.Data(0x32)
	d.Data(0x29)
	d.Data(0x2D)
	d.Data(0x29)
	d.Data(0x25)
	d.Data(0x2B)
	d.Data(0x39)
	d.Data(0x00)
	d.Data(0x01)
	d.Data(0x03)
	d.Data(0x10)
	d.Command(GMCTRN1)
	d.Data(0x03)
	d.Data(0x1D)
	d.Data(0x07)
	d.Data(0x06)
	d.Data(0x2E)
	d.Data(0x2C)
	d.Data(0x29)
	d.Data(0x2D)
	d.Data(0x2E)
	d.Data(0x2E)
	d.Data(0x37)
	d.Data(0x3F)
	d.Data(0x00)
	d.Data(0x00)
	d.Data(0x02)
	d.Data(0x10)

	d.Command(NORON)
	time.Sleep(10 * time.Millisecond)
	d.Command(DISPON)
	time.Sleep(500 * time.Millisecond)

	if cfg.Model == MINI80x160 {
		d.Command(MADCTL)
		d.Data(0xC0)
	}

	d.SetRotation(d.rotation)

	d.blPin.High()
}

// Display does nothing, there's no buffer as it might be too big for some boards
func (d *DeviceOf[T]) Display() error {
	return nil
}

// SetPixel sets a pixel in the screen
func (d *DeviceOf[T]) SetPixel(x int16, y int16, c color.RGBA) {
	w, h := d.Size()
	if x < 0 || y < 0 || x >= w || y >= h {
		return
	}
	d.FillRectangle(x, y, 1, 1, c)
}

// setWindow prepares the screen to be modified at a given rectangle
func (d *DeviceOf[T]) setWindow(x, y, w, h int16) {
	if d.rotation == drivers.Rotation0 || d.rotation == drivers.Rotation180 {
		x += d.columnOffset
		y += d.rowOffset
	} else {
		x += d.rowOffset
		y += d.columnOffset
	}
	d.Tx([]uint8{CASET}, true)
	d.Tx([]uint8{uint8(x >> 8), uint8(x), uint8((x + w - 1) >> 8), uint8(x + w - 1)}, false)
	d.Tx([]uint8{RASET}, true)
	d.Tx([]uint8{uint8(y >> 8), uint8(y), uint8((y + h - 1) >> 8), uint8(y + h - 1)}, false)
	d.Command(RAMWR)
}

// SetScrollWindow sets an area to scroll with fixed top and bottom parts of the display
func (d *DeviceOf[T]) SetScrollArea(topFixedArea, bottomFixedArea int16) {
	// TODO: this code is broken, see the st7789 and ili9341 implementations for
	// how to do this correctly.
	d.Command(VSCRDEF)
	d.Tx([]uint8{
		uint8(topFixedArea >> 8), uint8(topFixedArea),
		uint8(d.height - topFixedArea - bottomFixedArea>>8), uint8(d.height - topFixedArea - bottomFixedArea),
		uint8(bottomFixedArea >> 8), uint8(bottomFixedArea)},
		false)
}

// SetScroll sets the vertical scroll address of the display.
func (d *DeviceOf[T]) SetScroll(line int16) {
	d.Command(VSCRSADD)
	d.Tx([]uint8{uint8(line >> 8), uint8(line)}, false)
}

// SpotScroll returns the display to its normal state
func (d *DeviceOf[T]) StopScroll() {
	d.Command(NORON)
}

// FillRectangle fills a rectangle at a given coordinates with a color
func (d *DeviceOf[T]) FillRectangle(x, y, width, height int16, c color.RGBA) error {
	k, i := d.Size()
	if x < 0 || y < 0 || width <= 0 || height <= 0 ||
		x >= k || (x+width) > k || y >= i || (y+height) > i {
		return errors.New("rectangle coordinates outside display area")
	}
	d.setWindow(x, y, width, height)

	d.batchData.FillSolidColor(pixel.NewColor[T](c.R, c.G, c.B))
	i = width * height
	for i > 0 {
		if i >= d.batchLength {
			d.Tx(d.batchData.RawBuffer(), false)
		} else {
			d.Tx(d.batchData.Rescale(int(i), 1).RawBuffer(), false)
		}
		i -= d.batchLength
	}
	return nil
}

// DrawRGBBitmap8 copies an RGB bitmap to the internal buffer at given coordinates
//
// Deprecated: use DrawBitmap instead.
func (d *DeviceOf[T]) DrawRGBBitmap8(x, y int16, data []uint8, w, h int16) error {
	k, i := d.Size()
	if x < 0 || y < 0 || w <= 0 || h <= 0 ||
		x >= k || (x+w) > k || y >= i || (y+h) > i {
		return errOutOfBounds
	}
	d.setWindow(x, y, w, h)
	d.Tx(data, false)
	return nil
}

// DrawBitmap copies the bitmap to the internal buffer on the screen at the
// given coordinates. It returns once the image data has been sent completely.
func (d *DeviceOf[T]) DrawBitmap(x, y int16, bitmap pixel.Image[T]) error {
	width, height := bitmap.Size()
	return d.DrawRGBBitmap8(x, y, bitmap.RawBuffer(), int16(width), int16(height))
}

// FillRectangle fills a rectangle at a given coordinates with a buffer
func (d *DeviceOf[T]) FillRectangleWithBuffer(x, y, width, height int16, buffer []color.RGBA) error {
	k, l := d.Size()
	if x < 0 || y < 0 || width <= 0 || height <= 0 ||
		x >= k || (x+width) > k || y >= l || (y+height) > l {
		return errors.New("rectangle coordinates outside display area")
	}
	k = width * height
	l = int16(len(buffer))
	if k != l {
		return errors.New("buffer length does not match with rectangle size")
	}

	d.setWindow(x, y, width, height)

	offset := int16(0)
	for k > 0 {
		for i := int16(0); i < d.batchLength; i++ {
			if offset+i < l {
				c := buffer[offset+i]
				d.batchData.Set(int(i), 0, pixel.NewColor[T](c.R, c.G, c.B))
			}
		}
		if k >= d.batchLength {
			d.Tx(d.batchData.RawBuffer(), false)
		} else {
			d.Tx(d.batchData.Rescale(int(k), 1).RawBuffer(), false)
		}
		k -= d.batchLength
		offset += d.batchLength
	}
	return nil
}

// DrawFastVLine draws a vertical line faster than using SetPixel
func (d *DeviceOf[T]) DrawFastVLine(x, y0, y1 int16, c color.RGBA) {
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	d.FillRectangle(x, y0, 1, y1-y0+1, c)
}

// DrawFastHLine draws a horizontal line faster than using SetPixel
func (d *DeviceOf[T]) DrawFastHLine(x0, x1, y int16, c color.RGBA) {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	d.FillRectangle(x0, y, x1-x0+1, 1, c)
}

// FillScreen fills the screen with a given color
func (d *DeviceOf[T]) FillScreen(c color.RGBA) {
	if d.rotation == drivers.Rotation0 || d.rotation == drivers.Rotation180 {
		d.FillRectangle(0, 0, d.width, d.height, c)
	} else {
		d.FillRectangle(0, 0, d.height, d.width, c)
	}
}

// Rotation returns the currently configured rotation.
func (d *DeviceOf[T]) Rotation() drivers.Rotation {
	return d.rotation
}

// SetRotation changes the rotation of the device (clock-wise)
func (d *DeviceOf[T]) SetRotation(rotation drivers.Rotation) error {
	d.rotation = rotation
	madctl := uint8(0)
	switch rotation % 4 {
	case drivers.Rotation0:
		madctl = MADCTL_MX | MADCTL_MY
	case drivers.Rotation90:
		madctl = MADCTL_MY | MADCTL_MV
	case drivers.Rotation180:
		// nothing to do
	case drivers.Rotation270:
		madctl = MADCTL_MX | MADCTL_MV
	}
	if d.isBGR {
		madctl |= MADCTL_BGR
	}
	d.Command(MADCTL)
	d.Data(madctl)
	return nil
}

// Command sends a command to the display
func (d *DeviceOf[T]) Command(command uint8) {
	d.Tx([]byte{command}, true)
}

// Command sends a data to the display
func (d *DeviceOf[T]) Data(data uint8) {
	d.Tx([]byte{data}, false)
}

// Tx sends data to the display
func (d *DeviceOf[T]) Tx(data []byte, isCommand bool) {
	d.dcPin.Set(!isCommand)
	d.bus.Tx(data, nil)
}

// Size returns the current size of the display.
func (d *DeviceOf[T]) Size() (w, h int16) {
	if d.rotation == drivers.Rotation0 || d.rotation == drivers.Rotation180 {
		return d.width, d.height
	}
	return d.height, d.width
}

// EnableBacklight enables or disables the backlight
func (d *DeviceOf[T]) EnableBacklight(enable bool) {
	if enable {
		d.blPin.High()
	} else {
		d.blPin.Low()
	}
}

// Set the sleep mode for this LCD panel. When sleeping, the panel uses a lot
// less power. The LCD won't display an image anymore, but the memory contents
// will be kept.
func (d *DeviceOf[T]) Sleep(sleepEnabled bool) error {
	if sleepEnabled {
		// Shut down LCD panel.
		d.Command(SLPIN)
		time.Sleep(5 * time.Millisecond) // 5ms required by the datasheet
	} else {
		// Turn the LCD panel back on.
		d.Command(SLPOUT)
		// The st7735 datasheet says it is necessary to wait 120ms before
		// sending another command.
		time.Sleep(120 * time.Millisecond)
	}
	return nil
}

// InverColors inverts the colors of the screen
func (d *DeviceOf[T]) InvertColors(invert bool) {
	if invert {
		d.Command(INVON)
	} else {
		d.Command(INVOFF)
	}
}

// IsBGR changes the color mode (RGB/BGR)
func (d *DeviceOf[T]) IsBGR(bgr bool) {
	d.isBGR = bgr
}
