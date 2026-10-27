// Package ili9325 implements a driver for the ILI9325 TFT controller on an
// 8-bit parallel bus, as used on the ITead ITDB02-2.8 shield.
//
// Datasheet: https://cdn-shop.adafruit.com/datasheets/ILI9325.pdf
package ili9325 // import "tinygo.org/x/drivers/ili9325"

import (
	"errors"
	"image/color"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/pin"
	"tinygo.org/x/drivers/pixel"
)

// Bus is an 8-bit parallel data bus including its write strobe.
type Bus interface {
	// WriteBytes writes each byte in turn, pulsing WR after each one.
	WriteBytes(buf []byte)

	// WriteRepeat writes v high byte first, n times.
	WriteRepeat(v uint16, n int)
}

// Image buffer type used by the ili9325.
type Image = pixel.Image[pixel.RGB565BE]

const (
	width  = 240
	height = 320
)

var (
	errOutOfBounds = errors.New("ili9325: rectangle coordinates outside display area")
	errShortData   = errors.New("ili9325: pixel data shorter than rectangle")
)

// Config is the configuration for the display.
type Config struct {
	Rotation drivers.Rotation
}

// Device is an ILI9325 display on an 8-bit parallel bus.
type Device struct {
	bus      Bus
	rs       pin.OutputFunc
	cs       pin.OutputFunc
	rst      pin.OutputFunc
	rotation drivers.Rotation
	buf      [2]byte
}

// New returns a new ili9325 driver. The rs, cs and rst pins must already be
// configured as outputs.
func New(bus Bus, rs, cs, rst pin.Output) *Device {
	return &Device{
		bus: bus,
		rs:  rs.Set,
		cs:  cs.Set,
		rst: rst.Set,
	}
}

type reg struct {
	addr  uint8
	value uint16
	delay time.Duration
}

// Power on sequence and gamma from the UTFT library ILI9325D driver, see
// https://github.com/gnulabis/UTFT-ESP/blob/master/UTFT/tft_drivers/ili9325d/default/initlcd.h
var initSequence = [...]reg{
	{0xE5, 0x78F0, 0}, {DRIVER_OUTPUT_CONTROL, 0x0100, 0}, {LCD_DRIVING_CONTROL, 0x0200, 0},
	{ENTRY_MODE, entryModeBGR | entryModeIDX | entryModeIDY, 0},
	{RESIZE_CONTROL, 0x0000, 0}, {DISPLAY_CONTROL2, 0x0207, 0}, {DISPLAY_CONTROL3, 0x0000, 0},
	{DISPLAY_CONTROL4, 0x0000, 0}, {RGB_INTERFACE_CONTROL1, 0x0000, 0},
	{FRAME_MARKER_POSITION, 0x0000, 0}, {RGB_INTERFACE_CONTROL2, 0x0000, 0},
	{POWER_CONTROL1, 0x0000, 0}, {POWER_CONTROL2, 0x0007, 0}, {POWER_CONTROL3, 0x0000, 0},
	{POWER_CONTROL4, 0x0000, 0}, {DISPLAY_CONTROL1, 0x0001, 200 * time.Millisecond},
	{POWER_CONTROL1, 0x1690, 0}, {POWER_CONTROL2, 0x0227, 50 * time.Millisecond},
	{POWER_CONTROL3, 0x000D, 50 * time.Millisecond},
	{POWER_CONTROL4, 0x1200, 0}, {POWER_CONTROL7, 0x000A, 0},
	{FRAME_RATE_CONTROL, 0x000D, 50 * time.Millisecond},
	{GRAM_HORIZONTAL_ADDRESS, 0x0000, 0}, {GRAM_VERTICAL_ADDRESS, 0x0000, 0},
	{0x30, 0x0000, 0}, {0x31, 0x0404, 0}, {0x32, 0x0003, 0}, {0x35, 0x0405, 0},
	{0x36, 0x0808, 0}, {0x37, 0x0407, 0}, {0x38, 0x0303, 0}, {0x39, 0x0707, 0},
	{0x3C, 0x0504, 0}, {0x3D, 0x0808, 0},
	{HORIZONTAL_START, 0x0000, 0}, {HORIZONTAL_END, width - 1, 0},
	{VERTICAL_START, 0x0000, 0}, {VERTICAL_END, height - 1, 0},
	{GATE_SCAN_CONTROL1, 0xA700, 0}, {GATE_SCAN_CONTROL2, 0x0001, 0}, {GATE_SCAN_CONTROL3, 0x0000, 0},
	{0x80, 0x0000, 0}, {0x81, 0x0000, 0}, {0x82, 0x0000, 0}, {0x83, 0x0000, 0},
	{0x84, 0x0000, 0}, {0x85, 0x0000, 0},
	{PANEL_INTERFACE_CONTROL1, 0x0010, 0}, {PANEL_INTERFACE_CONTROL2, 0x0000, 0},
	{DISPLAY_CONTROL1, 0x0133, 0},
}

// Configure resets and initialises the display.
func (d *Device) Configure(config Config) error {
	d.cs.High()
	d.rst.High()
	time.Sleep(5 * time.Millisecond)
	d.rst.Low()
	time.Sleep(15 * time.Millisecond)
	d.rst.High()
	time.Sleep(15 * time.Millisecond)

	d.cs.Low()
	for _, r := range initSequence {
		d.writeReg(r.addr, r.value)
		if r.delay > 0 {
			time.Sleep(r.delay)
		}
	}
	d.cs.High()
	return d.SetRotation(config.Rotation)
}

// Size returns the current size of the display.
func (d *Device) Size() (x, y int16) {
	if d.rotation == drivers.Rotation90 || d.rotation == drivers.Rotation270 {
		return height, width
	}
	return width, height
}

// Rotation returns the current rotation of the display.
func (d *Device) Rotation() drivers.Rotation {
	return d.rotation
}

// SetRotation changes the rotation of the display (clockwise). Mirrored
// rotations are not supported.
func (d *Device) SetRotation(rotation drivers.Rotation) error {
	switch rotation {
	case drivers.Rotation0, drivers.Rotation90, drivers.Rotation180, drivers.Rotation270:
	default:
		return errors.New("ili9325: unsupported rotation")
	}
	d.rotation = rotation
	d.cs.Low()
	d.writeReg(ENTRY_MODE, entryModeBGR|d.entryMode())
	d.cs.High()
	return nil
}

// SetPixel draws a single pixel.
func (d *Device) SetPixel(x, y int16, c color.RGBA) {
	w, h := d.Size()
	if x < 0 || y < 0 || x >= w || y >= h {
		return
	}
	d.cs.Low()
	d.setWindow(x, y, 1, 1)
	d.bus.WriteRepeat(RGBATo565(c), 1)
	d.cs.High()
}

// Display does nothing, there is no buffer.
func (d *Device) Display() error {
	return nil
}

// FillRectangle fills a rectangle at the given coordinates with a colour.
func (d *Device) FillRectangle(x, y, w, h int16, c color.RGBA) error {
	if !d.inBounds(x, y, w, h) {
		return errOutOfBounds
	}
	d.cs.Low()
	d.setWindow(x, y, w, h)
	d.bus.WriteRepeat(RGBATo565(c), int(w)*int(h))
	d.cs.High()
	return nil
}

// FillScreen fills the whole screen with a colour.
func (d *Device) FillScreen(c color.RGBA) {
	w, h := d.Size()
	d.FillRectangle(0, 0, w, h, c)
}

// DrawRGBBitmap8 draws big endian RGB565 pixel data at the given coordinates.
func (d *Device) DrawRGBBitmap8(x, y int16, data []uint8, w, h int16) error {
	if !d.inBounds(x, y, w, h) {
		return errOutOfBounds
	}
	if len(data) < int(w)*int(h)*2 {
		return errShortData
	}
	d.cs.Low()
	d.setWindow(x, y, w, h)
	d.bus.WriteBytes(data[:int(w)*int(h)*2])
	d.cs.High()
	return nil
}

// DrawBitmap draws the bitmap at the given coordinates.
func (d *Device) DrawBitmap(x, y int16, bitmap Image) error {
	w, h := bitmap.Size()
	return d.DrawRGBBitmap8(x, y, bitmap.RawBuffer(), int16(w), int16(h))
}

func (d *Device) inBounds(x, y, w, h int16) bool {
	dw, dh := d.Size()
	return x >= 0 && y >= 0 && w > 0 && h > 0 && x+w <= dw && y+h <= dh
}

// entryMode returns the address counter direction that writes pixels in row
// order for the current rotation.
func (d *Device) entryMode() uint16 {
	switch d.rotation {
	case drivers.Rotation90:
		return entryModeIDY | entryModeAM
	case drivers.Rotation180:
		return 0
	case drivers.Rotation270:
		return entryModeIDX | entryModeAM
	default:
		return entryModeIDX | entryModeIDY
	}
}

// setWindow maps a window in rotated coordinates to the panel.
func (d *Device) setWindow(x, y, w, h int16) {
	var x0, x1, y0, y1, cx, cy int16
	switch d.rotation {
	case drivers.Rotation90:
		x0, x1 = width-y-h, width-1-y
		y0, y1 = x, x+w-1
		cx, cy = x1, y0
	case drivers.Rotation180:
		x0, x1 = width-x-w, width-1-x
		y0, y1 = height-y-h, height-1-y
		cx, cy = x1, y1
	case drivers.Rotation270:
		x0, x1 = y, y+h-1
		y0, y1 = height-x-w, height-1-x
		cx, cy = x0, y1
	default:
		x0, x1 = x, x+w-1
		y0, y1 = y, y+h-1
		cx, cy = x0, y0
	}
	d.writeReg(HORIZONTAL_START, uint16(x0))
	d.writeReg(HORIZONTAL_END, uint16(x1))
	d.writeReg(VERTICAL_START, uint16(y0))
	d.writeReg(VERTICAL_END, uint16(y1))
	d.writeReg(GRAM_HORIZONTAL_ADDRESS, uint16(cx))
	d.writeReg(GRAM_VERTICAL_ADDRESS, uint16(cy))
	d.writeIndex(WRITE_DATA_TO_GRAM)
	d.rs.High()
}

func (d *Device) writeIndex(addr uint8) {
	d.rs.Low()
	d.buf = [2]byte{0, addr}
	d.bus.WriteBytes(d.buf[:])
}

func (d *Device) writeReg(addr uint8, value uint16) {
	d.writeIndex(addr)
	d.rs.High()
	d.bus.WriteRepeat(value, 1)
}

// RGBATo565 converts a color.RGBA to the RGB565 value used by the display.
func RGBATo565(c color.RGBA) uint16 {
	return uint16(c.R&0xF8)<<8 | uint16(c.G&0xFC)<<3 | uint16(c.B)>>3
}
