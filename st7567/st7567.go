// Package st7567 implements a driver for the ST7567 and AiP31567 monochrome LCD
// controllers over 4-line SPI.
//
// Datasheet: https://github.com/WeActStudio/WeActStudio.MatrixLCDModule/blob/master/Doc/ST_7567_V1_5.pdf
package st7567 // import "tinygo.org/x/drivers/st7567"

import (
	"errors"
	"image/color"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/pin"
)

var (
	errUnsupportedRotation = errors.New("st7567: rotation must be Rotation0 or Rotation180")
	errOutOfRange          = errors.New("st7567: out of range")
)

// Config is the configuration for the display.
type Config struct {
	// Width and Height of the panel in pixels. The default is 128x64.
	Width  int16
	Height int16

	// Rotation is Rotation0 or Rotation180. Rotation0 is upright on the WeAct
	// Studio 1.37 inch module.
	Rotation drivers.Rotation

	// RegulationRatio selects the V0 regulator ratio, 1 to 7 for 3.5 to 6.5.
	// Zero selects the default, 4 for 5.0.
	RegulationRatio uint8

	// Contrast is the electronic volume, 1 to 63. Zero selects the default.
	Contrast uint8
}

// Device is an ST7567 display connected over SPI.
type Device struct {
	bus      drivers.SPI
	cs       pin.OutputFunc
	dc       pin.OutputFunc
	rst      pin.OutputFunc
	width    int16
	height   int16
	rotation drivers.Rotation
	buffer   []byte
	cmd      [2]byte
}

// New returns a new ST7567 driver. The SPI bus and the pins must already be
// configured, with all pins as outputs.
func New(bus drivers.SPI, csPin, dcPin, resetPin pin.Output) *Device {
	return &Device{
		bus: bus,
		cs:  csPin.Set,
		dc:  dcPin.Set,
		rst: resetPin.Set,
	}
}

// Configure resets and initializes the display, then clears it.
func (d *Device) Configure(cfg Config) error {
	d.width = cfg.Width
	if d.width == 0 {
		d.width = 128
	}
	d.height = cfg.Height
	if d.height == 0 {
		d.height = 64
	}
	if d.width > maxColumns || d.height > 64 || d.height%8 != 0 {
		return errOutOfRange
	}
	regulation := cfg.RegulationRatio
	if regulation == 0 {
		regulation = defaultRegulation
	}
	contrast := cfg.Contrast
	if contrast == 0 {
		contrast = defaultContrast
	}
	if regulation > 7 || contrast > maxContrast {
		return errOutOfRange
	}
	d.buffer = make([]byte, int(d.width)*int(d.height)/8)

	d.cs.High()
	d.rst.Low()
	time.Sleep(10 * time.Microsecond)
	d.rst.High()
	time.Sleep(10 * time.Microsecond)

	// Power control must follow the reset within 5ms, see "Initializing with
	// the Built-in Power Supply Circuits" in the datasheet.
	d.command(cmdBias9)
	if err := d.SetRotation(cfg.Rotation); err != nil {
		return err
	}
	d.command(cmdRegulation | regulation)
	d.command2(cmdSetEV, contrast)
	d.command(cmdPowerControl | powerBoosterOn | powerRegulatorOn | powerFollowerOn)
	d.command(cmdStartLine)
	d.command(cmdAllPixelOff)
	d.command(cmdInverseOff)

	if err := d.Display(); err != nil {
		return err
	}
	d.command(cmdDisplayOn)
	return nil
}

// SetRotation sets Rotation0 or Rotation180. Call Display to redraw.
func (d *Device) SetRotation(rotation drivers.Rotation) error {
	switch rotation {
	case drivers.Rotation0:
		d.command(cmdSEGReverse)
		d.command(cmdCOMNormal)
	case drivers.Rotation180:
		d.command(cmdSEGNormal)
		d.command(cmdCOMReverse)
	default:
		return errUnsupportedRotation
	}
	d.rotation = rotation
	return nil
}

// Rotation returns the current rotation.
func (d *Device) Rotation() drivers.Rotation {
	return d.rotation
}

// SetContrast sets the electronic volume, 0 to 63.
func (d *Device) SetContrast(contrast uint8) {
	if contrast > maxContrast {
		contrast = maxContrast
	}
	d.command2(cmdSetEV, contrast)
}

// SetInverted inverts all pixels on the display without changing the buffer.
func (d *Device) SetInverted(inverted bool) {
	if inverted {
		d.command(cmdInverseOn)
	} else {
		d.command(cmdInverseOff)
	}
}

// Sleep turns the display off when sleep is true and on again when false.
func (d *Device) Sleep(sleep bool) error {
	if sleep {
		d.command(cmdDisplayOff)
	} else {
		d.command(cmdDisplayOn)
	}
	return nil
}

// Size returns the display size in pixels.
func (d *Device) Size() (x, y int16) {
	return d.width, d.height
}

// GetBuffer returns the buffer. Each byte holds 8 vertical pixels with the top
// pixel in bit 0, and each page of 8 rows is Width bytes.
func (d *Device) GetBuffer() []byte {
	return d.buffer
}

// ClearBuffer clears the buffer.
func (d *Device) ClearBuffer() {
	clear(d.buffer)
}

// ClearDisplay clears the buffer and the display.
func (d *Device) ClearDisplay() error {
	d.ClearBuffer()
	return d.Display()
}

// SetPixel sets a pixel in the buffer. Any colour other than black turns the
// pixel on.
func (d *Device) SetPixel(x, y int16, c color.RGBA) {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return
	}
	i := int(x) + int(y/8)*int(d.width)
	if c.R != 0 || c.G != 0 || c.B != 0 {
		d.buffer[i] |= 1 << uint(y%8)
	} else {
		d.buffer[i] &^= 1 << uint(y%8)
	}
}

// GetPixel reports whether a pixel in the buffer is on.
func (d *Device) GetPixel(x, y int16) bool {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return false
	}
	return d.buffer[int(x)+int(y/8)*int(d.width)]&(1<<uint(y%8)) != 0
}

// Display sends the buffer to the display.
func (d *Device) Display() error {
	// With SEG reversed, column 0 drives SEG131, so skip the unused columns.
	offset := uint8(0)
	if d.rotation == drivers.Rotation0 {
		offset = maxColumns - uint8(d.width)
	}
	w := int(d.width)
	for page := 0; page < int(d.height)/8; page++ {
		d.command(cmdPageAddress | uint8(page))
		d.command(cmdColumnHigh | offset>>4)
		d.command(cmdColumnLow | offset&0x0F)
		if err := d.data(d.buffer[page*w : (page+1)*w]); err != nil {
			return err
		}
	}
	return nil
}

func (d *Device) command(c uint8) error {
	d.cmd[0] = c
	return d.write(d.cmd[:1], false)
}

func (d *Device) command2(c, v uint8) error {
	d.cmd[0] = c
	d.cmd[1] = v
	return d.write(d.cmd[:2], false)
}

func (d *Device) data(b []byte) error {
	return d.write(b, true)
}

func (d *Device) write(b []byte, isData bool) error {
	d.dc(isData)
	d.cs.Low()
	err := d.bus.Tx(b, nil)
	d.cs.High()
	return err
}
