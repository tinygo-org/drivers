// Package gc9a01 implements a driver for the gc9a01 LCD round display
//
// Datasheet: https://www.waveshare.com/w/upload/5/5e/GC9A01A.pdf
package gc9a01 // import "tinygo.org/x/drivers/gc9a01"

import (
	"image/color"
	"machine"
	"time"

	"errors"

	"tinygo.org/x/drivers"
)

// Rotation controls the rotation used by the display.
type Orientation uint8

// FrameRate controls the frame rate used by the display.
type FrameRate uint8

// Device wraps an SPI connection.
type Device struct {
	bus             drivers.SPI
	dcPin           machine.Pin
	resetPin        machine.Pin
	csPin           machine.Pin
	blPin           machine.Pin
	width           int16
	height          int16
	columnOffsetCfg int16
	rowOffsetCfg    int16
	columnOffset    int16
	rowOffset       int16
	frameRate       FrameRate
	isBGR           bool
	vSyncLines      int16
	orientation     Orientation
	batchLength     int16
	batchData       []uint8
}

// Config is the configuration for the display
type Config struct {
	Orientation  Orientation
	RowOffset    int16
	ColumnOffset int16
	FrameRate    FrameRate
	VSyncLines   int16
	Width        int16
	Height       int16
}

// New creates a new ST7789 connection. The SPI wire must already be configured.
func New(bus drivers.SPI, resetPin, dcPin, csPin, blPin machine.Pin) Device {
	resetPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	dcPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	blPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return Device{
		bus:      bus,
		resetPin: resetPin,
		dcPin:    dcPin,
		csPin:    csPin,
		blPin:    blPin,
	}
}

// Reset the Device
func (d *Device) Reset() {
	d.resetPin.High()
	time.Sleep(100 * time.Millisecond)
	d.resetPin.Low()
	time.Sleep(100 * time.Millisecond)
	d.resetPin.High()
	time.Sleep(100 * time.Millisecond)

}

// Sets Device configuration for screen orientation using the initialzation values
func (d *Device) SetDeviceOrientation() {

	var MemoryAccessReg uint8

	//Get GRAM and LCD width and height
	if d.orientation == HORIZONTAL {
		MemoryAccessReg = 0xc8
	} else {
		MemoryAccessReg = 0x68
	}

	// Set the read / write scan direction of the frame memory
	d.Command(MADCTR)
	//0x08 set RGB
	d.Command(MemoryAccessReg)
}

// setWindow prepares the screen to be modified at a given rectangle
func (d *Device) setWindow(x, y, w, h int16) {
	if d.orientation == HORIZONTAL {
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

// FillScreen fills the screen with a given color
func (d *Device) FillScreen(c color.RGBA) {
	d.FillRectangle(0, 0, d.height, d.width, c)
}

// FillRectangle fills a rectangle at a given coordinates with a color
func (d *Device) FillRectangle(x, y, width, height int16, c color.RGBA) error {
	k, j := d.Size()
	var i int32
	if x < 0 || y < 0 || width <= 0 || height <= 0 ||
		x >= k || (x+width) > k || y >= j || (y+height) > j {
		return errors.New("rectangle coordinates outside display area")
	}
	d.setWindow(x, y, width, height)
	c565 := RGBATo565(c)
	c1 := uint8(c565 >> 8)
	c2 := uint8(c565)

	for i = 0; i < int32(d.batchLength); i++ {
		d.batchData[i*2] = c1
		d.batchData[i*2+1] = c2
	}
	i = int32(width) * int32(height)
	batchLength := int32(d.batchLength)
	for i > 0 {
		if i >= batchLength {
			d.Tx(d.batchData, false)
		} else {
			d.Tx(d.batchData[:i*2], false)
		}
		i -= batchLength
	}
	return nil
}

// Display sends the whole buffer to the screen
func (d *Device) Display() error {
	return nil
}

// FillRectangleWithBuffer fills buffer with a rectangle at a given coordinates.
func (d *Device) FillRectangleWithBuffer(x, y, width, height int16, buffer []color.RGBA) error {
	h, w := d.Size()
	if x < 0 || y < 0 || width <= 0 || height <= 0 ||
		x >= h || (x+width) > h || y >= w || (y+height) > w {
		return errors.New("rectangle coordinates outside display area")
	}
	k := int32(width) * int32(height)
	l := int32(len(buffer))
	if k != l {
		return errors.New("buffer length does not match with rectangle size")
	}

	d.setWindow(x, y, width, height)

	offset := int32(0)
	batchLength := int32(d.batchLength)
	for k > 0 {
		for i := int32(0); i < batchLength; i++ {
			if offset+i < l {
				c565 := RGBATo565(buffer[offset+i])
				c1 := uint8(c565 >> 8)
				c2 := uint8(c565)
				d.batchData[i*2] = c1
				d.batchData[i*2+1] = c2
			}
		}
		if k >= batchLength {
			d.Tx(d.batchData, false)
		} else {
			d.Tx(d.batchData[:k*2], false)
		}
		k -= batchLength
		offset += batchLength
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

// SetPixel sets a pixel in the screen
func (d *Device) SetPixel(x, y int16, c color.RGBA) {
	w, h := d.Size()
	if x < 0 || y < 0 || x >= w || y >= h {
		return
	}
	d.FillRectangle(x, y, 1, 1, c)
}

// Command sends a command to the display.
func (d *Device) Command(command uint8) {
	d.Tx([]byte{command}, true)
}

// Data sends data to the display.
func (d *Device) Data(data uint8) {
	d.Tx([]byte{data}, false)
}

// Tx sends data to the display
func (d *Device) Tx(data []byte, isCommand bool) {
	d.dcPin.Set(!isCommand)
	d.bus.Tx(data, nil)
}

// Rx reads data from the display
func (d *Device) Rx(command uint8, data []byte) {
	d.dcPin.Low()
	d.csPin.Low()
	d.bus.Transfer(command)
	d.dcPin.High()
	for i := range data {
		data[i], _ = d.bus.Transfer(0xFF)
	}
	d.csPin.High()
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return d.height, d.width
}

// EnableBacklight enables or disables the backlight
func (d *Device) EnableBacklight(enable bool) {
	if enable {
		d.blPin.High()
	} else {
		d.blPin.Low()
	}
}

// InvertColors inverts the colors of the screen
func (d *Device) InvertColors(invert bool) {
	if invert {
		d.Command(INVON)
	} else {
		d.Command(INVOFF)
	}
}

// IsBGR changes the color mode (RGB/BGR)
func (d *Device) IsBGR(bgr bool) {
	d.isBGR = bgr
}

// SetScrollArea sets an area to scroll with fixed top and bottom parts of the display.
func (d *Device) SetScrollArea(topFixedArea, bottomFixedArea int16) {
	d.Command(VSCRDEF)
	d.Tx([]uint8{
		uint8(topFixedArea >> 8), uint8(topFixedArea),
		uint8(d.height - topFixedArea - bottomFixedArea>>8), uint8(d.height - topFixedArea - bottomFixedArea),
		uint8(bottomFixedArea >> 8), uint8(bottomFixedArea)},
		false)
}

// SetScroll sets the vertical scroll address of the display.
func (d *Device) SetScroll(line int16) {
	d.Command(VSCRSADD)
	d.Tx([]uint8{uint8(line >> 8), uint8(line)}, false)
}

// StopScroll returns the display to its normal state.
func (d *Device) StopScroll() {
	d.Command(NORON)
}

// RGBATo565 converts a color.RGBA to uint16 used in the display
func RGBATo565(c color.RGBA) uint16 {
	r, g, b, _ := c.RGBA()
	return uint16((r & 0xF800) +
		((g & 0xFC00) >> 5) +
		((b & 0xF800) >> 11))
}

// Configure initializes the display with default configuration
func (d *Device) Configure(cfg Config) {
	if cfg.Width != 0 {
		d.width = cfg.Width
	} else {
		d.width = 240
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	} else {
		d.height = 240
	}

	d.orientation = cfg.Orientation
	d.rowOffsetCfg = cfg.RowOffset
	d.columnOffsetCfg = cfg.ColumnOffset
	d.batchLength = d.width

	if cfg.VSyncLines >= 2 && cfg.VSyncLines <= 254 {
		d.vSyncLines = cfg.VSyncLines
	} else {
		d.vSyncLines = 16
	}
	d.batchLength += d.batchLength & 1
	d.batchData = make([]uint8, d.batchLength*2)

	// Reset the device
	d.Reset()

	// Set Device Attributes
	d.SetDeviceOrientation()

	// Common initialization
	d.Command(0xEF)
	d.Command(0xEB)
	d.Data(0x14)

	d.Command(INTEN1)
	d.Command(0xEF)

	d.Command(0xEB)
	d.Data(0x14)

	d.Command(0x84)
	d.Data(0x40)

	d.Command(0x85)
	d.Data(0xFF)

	d.Command(0x86)
	d.Data(0xFF)

	d.Command(0x87)
	d.Data(0xFF)

	d.Command(0x88)
	d.Data(0x0A)

	d.Command(0x89)
	d.Data(0x21)

	d.Command(0x8A)
	d.Data(0x00)

	d.Command(0x8B)
	d.Data(0x80)

	d.Command(0x8C)
	d.Data(0x01)

	d.Command(0x8D)
	d.Data(0x01)

	d.Command(0x8E)
	d.Data(0xFF)

	d.Command(0x8F)
	d.Data(0xFF)

	d.Command(DISFNCTL)
	d.Data(0x00)
	d.Data(0x20)

	d.Command(MADCTR)
	d.Data(0x08) //Set as vertical screen

	d.Command(COLMOD)
	d.Data(0x05)

	d.Command(0x90)
	d.Data(0x08)
	d.Data(0x08)
	d.Data(0x08)
	d.Data(0x08)

	d.Command(0xBD)
	d.Data(0x06)

	d.Command(0xBC)
	d.Data(0x00)

	d.Command(0xFF)
	d.Data(0x60)
	d.Data(0x01)
	d.Data(0x04)

	d.Command(PWCTR3)
	d.Data(0x13)
	d.Command(PWCTR4)
	d.Data(0x13)

	d.Command(0xC9)
	d.Data(0x22)

	d.Command(0xBE)
	d.Data(0x11)

	d.Command(0xE1)
	d.Data(0x10)
	d.Data(0x0E)

	d.Command(0xDF)
	d.Data(0x21)
	d.Data(0x0c)
	d.Data(0x02)

	d.Command(GMSET1)
	d.Data(0x45)
	d.Data(0x09)
	d.Data(0x08)
	d.Data(0x08)
	d.Data(0x26)
	d.Data(0x2A)

	d.Command(GMSET2)
	d.Data(0x43)
	d.Data(0x70)
	d.Data(0x72)
	d.Data(0x36)
	d.Data(0x37)
	d.Data(0x6F)

	d.Command(GMSET3)
	d.Data(0x45)
	d.Data(0x09)
	d.Data(0x08)
	d.Data(0x08)
	d.Data(0x26)
	d.Data(0x2A)

	d.Command(GMSET4)
	d.Data(0x43)
	d.Data(0x70)
	d.Data(0x72)
	d.Data(0x36)
	d.Data(0x37)
	d.Data(0x6F)

	d.Command(0xED)
	d.Data(0x1B)
	d.Data(0x0B)

	d.Command(0xAE)
	d.Data(0x77)

	d.Command(0xCD)
	d.Data(0x63)

	d.Command(0x70)
	d.Data(0x07)
	d.Data(0x07)
	d.Data(0x04)
	d.Data(0x0E)
	d.Data(0x0F)
	d.Data(0x09)
	d.Data(0x07)
	d.Data(0x08)
	d.Data(0x03)

	d.Command(FRMCTL)
	d.Data(0x34)

	d.Command(0x62)
	d.Data(0x18)
	d.Data(0x0D)
	d.Data(0x71)
	d.Data(0xED)
	d.Data(0x70)
	d.Data(0x70)
	d.Data(0x18)
	d.Data(0x0F)
	d.Data(0x71)
	d.Data(0xEF)
	d.Data(0x70)
	d.Data(0x70)

	d.Command(0x63)
	d.Data(0x18)
	d.Data(0x11)
	d.Data(0x71)
	d.Data(0xF1)
	d.Data(0x70)
	d.Data(0x70)
	d.Data(0x18)
	d.Data(0x13)
	d.Data(0x71)
	d.Data(0xF3)
	d.Data(0x70)
	d.Data(0x70)

	d.Command(0x64)
	d.Data(0x28)
	d.Data(0x29)
	d.Data(0xF1)
	d.Data(0x01)
	d.Data(0xF1)
	d.Data(0x00)
	d.Data(0x07)

	d.Command(0x66)
	d.Data(0x3C)
	d.Data(0x00)
	d.Data(0xCD)
	d.Data(0x67)
	d.Data(0x45)
	d.Data(0x45)
	d.Data(0x10)
	d.Data(0x00)
	d.Data(0x00)
	d.Data(0x00)

	d.Command(0x67)
	d.Data(0x00)
	d.Data(0x3C)
	d.Data(0x00)
	d.Data(0x00)
	d.Data(0x00)
	d.Data(0x01)
	d.Data(0x54)
	d.Data(0x10)
	d.Data(0x32)
	d.Data(0x98)

	d.Command(0x74)
	d.Data(0x10)
	d.Data(0x85)
	d.Data(0x80)
	d.Data(0x00)
	d.Data(0x00)
	d.Data(0x4E)
	d.Data(0x00)

	d.Command(0x98)
	d.Data(0x3e)
	d.Data(0x07)

	d.Command(TEON)
	d.Command(0x21)

	d.Command(SLPOUT)
	time.Sleep(120 * time.Millisecond)
	d.Command(DISPON)
	time.Sleep(20 * time.Millisecond)

	d.blPin.High()
}
