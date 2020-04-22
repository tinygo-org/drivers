// Package ssd1351 implements a driver for the SSD1351 OLED color displays.
//
// Datasheet: https://download.mikroe.com/documents/datasheets/ssd1351-revision-1.3.pdf
//
package ssd1351 // import "tinygo.org/x/drivers/ssd1351"

import (
	"errors"
	"image/color"
	"machine"
	"time"
)

// Device wraps an SPI connection.
type Device struct {
	bus          machine.SPI
	dcPin        machine.Pin
	resetPin     machine.Pin
	csPin        machine.Pin
	enPin        machine.Pin
	rwPin        machine.Pin
	width        int16
	height       int16
	bufferLength int16
	bufferData   []uint8
}

// Config is the configuration for the display
type Config struct {
	Width  int16
	Height int16
}

// New creates a new SSD1331 connection. The SPI wire must already be configured.
func New(bus machine.SPI, resetPin, dcPin, csPin, enPin, rwPin machine.Pin) Device {
	return Device{
		bus:      bus,
		dcPin:    dcPin,
		resetPin: resetPin,
		csPin:    csPin,
		enPin:    enPin,
		rwPin:    rwPin,
	}
}

// Configure initializes the display with default configuration
func (d *Device) Configure(cfg Config) {
	if cfg.Width != 0 {
		d.width = cfg.Width
	} else {
		d.width = 128
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	} else {
		d.height = 128
	}
	d.bufferLength = d.width
	if d.height > d.width {
		d.bufferLength = d.height
	}
	//d.bufferLength += d.bufferLength & 1
	d.bufferData = make([]uint8, d.bufferLength*2)
	print("config")

	// configure pins
	d.dcPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.resetPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.enPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.rwPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// reset the device
	d.resetPin.High()
	time.Sleep(100 * time.Millisecond)
	d.resetPin.Low()
	time.Sleep(100 * time.Millisecond)
	d.resetPin.High()
	time.Sleep(200 * time.Millisecond)

	d.rwPin.Low()
	d.dcPin.Low()
	d.enPin.High()

	// Initialization
	d.Command(SET_COMMAND_LOCK)
	d.Data(0x12)
	d.Command(SET_COMMAND_LOCK)
	d.Data(0xB1)
	d.Command(SLEEP_MODE_DISPLAY_OFF)
	d.Command(SET_FRONT_CLOCK_DIV)
	d.Data(0xF1)
	d.Command(SET_MUX_RATIO)
	d.Data(0x7F)
	d.Command(SET_REMAP_COLORDEPTH)
	d.Data(0x72)
	d.Command(SET_COLUMN_ADDRESS)
	d.Data(0x00)
	d.Data(0x7F)
	d.Command(SET_ROW_ADDRESS)
	d.Data(0x00)
	d.Data(0x7F)
	d.Command(SET_DISPLAY_START_LINE)
	d.Data(0x00)
	d.Command(SET_DISPLAY_OFFSET)
	d.Data(0x00)
	d.Command(SET_GPIO)
	d.Data(0x00)
	d.Command(FUNCTION_SELECTION)
	d.Data(0x01)
	d.Command(SET_PHASE_PERIOD)
	d.Data(0x32)
	d.Command(SET_SEGMENT_LOW_VOLTAGE)
	d.Data(0xA0)
	d.Data(0xB5)
	d.Data(0x55)
	d.Command(SET_PRECHARGE_VOLTAGE)
	d.Data(0x17)
	d.Command(SET_VCOMH_VOLTAGE)
	d.Data(0x05)
	d.Command(SET_CONTRAST)
	d.Data(0xC8)
	d.Data(0x80)
	d.Data(0xC8)
	d.Command(MASTER_CONTRAST)
	d.Data(0x0F)
	d.Command(SET_SECOND_PRECHARGE_PERIOD)
	d.Data(0x01)
	d.Command(SET_DISPLAY_MODE_RESET)
	d.Command(SLEEP_MODE_DISPLAY_ON)

}

// Display does nothing, there's no buffer as it might be too big for some boards
func (d *Device) Display() error {
	return nil
}

// SetPixel sets a pixel in the buffer
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || y < 0 || x >= d.width || y >= d.height {
		return
	}
	/*c565 := RGBATo565(c)
	c1 := uint8(c565 >> 8)
	c2 := uint8(c565)
	index := y*d.width + 2*x
	d.bufferData[index] = pixel{
		Hword: c2,
		Lword: c1,
	}*/
	//d.Tx(d.bufferData, false)
	d.FillRectangle(x, y, 1, 1, c)
}

// setWindow prepares the screen to be modified at a given rectangle
func (d *Device) setWindow(x, y, w, h int16) {
	d.Command(SET_COLUMN_ADDRESS)
	d.Data(uint8(OLED_COL_OFFSET + x))
	d.Data(uint8(OLED_COL_OFFSET + x + w - 1))
	d.Command(SET_ROW_ADDRESS)
	d.Data(uint8(OLED_ROW_OFFSET + y))
	d.Data(uint8(OLED_ROW_OFFSET + y + h - 1))
	d.Command(WRITE_RAM)
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
	for i = 0; i < d.bufferLength; i++ {
		d.bufferData[i*2] = c1
		d.bufferData[i*2+1] = c2
	}
	i = width * height
	for i > 0 {
		if i >= d.bufferLength {
			d.Tx(d.bufferData, false)
		} else {
			d.Tx(d.bufferData[:i*2], false)
		}
		i -= d.bufferLength
	}

	return nil
}

// FillRectangleWithBuffer fills a rectangle at a given coordinates with a buffer
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
		for i := int16(0); i < d.bufferLength; i++ {
			if offset+i < l {
				c565 := RGBATo565(buffer[offset+i])
				c1 := uint8(c565 >> 8)
				c2 := uint8(c565)
				d.bufferData[i*2] = c1
				d.bufferData[i*2+1] = c2
			}
		}
		if k >= d.bufferLength {
			d.Tx(d.bufferData, false)
		} else {
			d.Tx(d.bufferData[:k*2], false)
		}
		k -= d.bufferLength
		offset += d.bufferLength
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
	d.FillRectangle(x0, y, x1-x0+1, y, c)
}

// FillScreen fills the screen with a given color
func (d *Device) FillScreen(c color.RGBA) {
	d.FillRectangle(0, 0, d.width, d.height, c)
}

// SetContrast sets the three contrast values (A, B & C)
func (d *Device) SetContrast(contrastA, contrastB, contrastC uint8) {
	d.Command(SET_CONTRAST)
	d.Data(contrastA)
	d.Data(contrastB)
	d.Data(contrastC)
}

// Command sends a command byte to the display
func (d *Device) Command(command uint8) {
	d.Tx([]byte{command}, true)
}

// Data sends a data byte to the display
func (d *Device) Data(data uint8) {
	d.Tx([]byte{data}, false)
}

// Tx sends data to the display
func (d *Device) Tx(data []byte, isCommand bool) {
	d.dcPin.Set(!isCommand)
	d.csPin.Low()
	d.bus.Tx(data, nil)
	d.dcPin.High()
	d.csPin.High()
}

// Size returns the current size of the display
func (d *Device) Size() (w, h int16) {
	return d.width, d.height
}

// RGBATo565 converts a color.RGBA to uint16 used in the display
func RGBATo565(c color.RGBA) uint16 {
	r, g, b, _ := c.RGBA()
	return uint16((r & 0xF800) +
		((g & 0xFC00) >> 5) +
		((b & 0xF800) >> 11))
}
