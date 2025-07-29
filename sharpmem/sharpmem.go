package sharpmem

import (
	"errors"
	"image/color"

	"tinygo.org/x/drivers"
)

const (
	bitWriteCmd uint8 = 0b00000001
	bitVcom     uint8 = 0b00000010
	bitClear    uint8 = 0b00000100
)

var (
	ConfigLS010B7DH04  = Config{Width: 128, Height: 128}
	ConfigLS011B7DH03  = Config{Width: 160, Height: 68}
	ConfigLS012B7DD01  = Config{Width: 184, Height: 38}
	ConfigLS013B7DH03  = ConfigLS010B7DH04
	ConfigLS013B7DH05  = Config{Width: 144, Height: 168}
	ConfigLS018B7DH02  = Config{Width: 230, Height: 303}
	ConfigLS027B7DH01  = Config{Width: 400, Height: 240}
	ConfigLS027B7DH01A = ConfigLS027B7DH01
	ConfigLS032B7DD02  = Config{Width: 336, Height: 536}
	ConfigLS044Q7DH01  = Config{Width: 320, Height: 240}
)

type Pin interface {
	High()
	Low()
}

// Device represents a Sharp Memory Display device. This driver implementation
// concerns the 1-bit color versions only (black and white memory displays).
//
// Supported SKUs include:
// LS010B7DH04, LS011B7DH03, LS012B7DD01, LS013B7DH03, LS013B7DH05,
// LS018B7DH02, LS027B7DH01, LS027B7DH01A, LS032B7DD02, LS044Q7DH01
//
// Note: Only SKU LS011B7DH03 (160x68) has been tested as of writing.
//
// The driver includes optimizations (frame and per-line invalidation) that
// only transmit the changed lines to the display. These optimizations are on
// by default, and they can be disabled with the respective config option.
type Device struct {
	bus          drivers.SPI
	csPin        Pin
	buffer       []byte
	txBuf        []byte
	lineDiff     []byte
	width        int16
	height       int16
	bufferSize   int16
	bytesPerLine int16
	vcom         uint8
	diffing      bool
}

type Config struct {
	Width  int16
	Height int16

	// DisableOptimizations disables frame and line invalidation optimizations.
	// Useful if constant frame times are desired.
	DisableOptimizations bool
}

// New creates a new device connection.
// The SPI bus must have already been configured.
func New(bus drivers.SPI, csPin Pin) Device {
	d := Device{
		bus:   bus,
		csPin: csPin,
	}
	return d
}

// Configure initializes the display with specified configuration. It can be
// called multiple times on the same display, resetting its internal state.
func (d *Device) Configure(cfg Config) {
	if cfg.Width == 0 {
		cfg.Width = 160
	}
	if cfg.Height == 0 {
		cfg.Height = 68
	}

	d.width = cfg.Width
	d.height = cfg.Height
	d.diffing = !cfg.DisableOptimizations

	d.initialize()
}

// initialize properly initializes the display and the in-memory image buffers.
func (d *Device) initialize() {
	d.csPin.Low()

	// initialize VCOM as high
	d.vcom = bitVcom

	// bytesPerLine has to be 16-bit aligned, as some resolutions require
	// padding to the nearest 2nd byte.
	d.bytesPerLine = ceilDiv(d.width, 16) * 2

	// preallocate a contiguous byte buffer for all lines, including
	// protocol-required padding for each line apriori (easier to transfer).
	d.bufferSize = d.bytesPerLine * d.height
	d.buffer = make([]byte, d.bufferSize)
	// A bit being 1 is white (reflective), 0 is black (less reflective).
	for i := range d.buffer {
		d.buffer[i] = 0xff
	}

	// auxiliary buffer for SPI transfers to avoid dynamic allocations
	d.txBuf = make([]byte, 2)

	if d.diffing {
		// buffer to store the changed lines. First bit is whether any line has
		// changed at all (i.e. the frame is invalid), followed by N bits,
		// one for each line.
		d.lineDiff = make([]byte, bitfieldBufLen(1+d.height))
	}
}

// SetPixel enables or disables a pixel in the buffer.
// color.RGBA{0, 0, 0, 255} is considered transparent (reflective, white),
// anything else will enable a pixel on the screen (make it appear less
// reflective, black).
func (d *Device) SetPixel(x, y int16, c color.RGBA) {
	if d.width == 0 {
		return
	}

	// bounds check
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return
	}

	offset := y * d.bytesPerLine

	div := offset + x/8
	mod := uint8(x % 8)

	prev := hasBit(d.buffer[div], mod)
	curr := c.R == 0 && c.G == 0 && c.B == 0 && c.A == 255

	if prev == curr {
		return
	}

	if curr {
		d.buffer[div] = setBit(d.buffer[div], mod)
	} else {
		d.buffer[div] = unsetBit(d.buffer[div], mod)
	}

	if d.diffing {
		d.invalidateLine(y)
	}
}

// Size returns the current size of the display.
func (d *Device) Size() (x, y int16) {
	return d.width, d.height
}

// Display renders the buffer to the screen. It only transmits changed lines if
// optimizations are enabled. It should be called at >=1hz, even if the
// buffer hasn't been modified.
func (d *Device) Display() error {
	if d.width == 0 {
		return errors.New("display not configured")
	}

	if d.diffing {
		if !hasBit(d.lineDiff[0], 0) {
			// no pixels have been modified, simply toggle VCOM
			return d.holdDisplay()
		}

		defer func() {
			for i := 0; i < len(d.lineDiff); i++ {
				d.lineDiff[i] = 0x00
			}
		}()
	}

	cmd := bitWriteCmd | d.vcom

	d.toggleVcom()

	// Padding to use for high bits of line numbers that overflow 8 bits.
	var hiPad = uint8(0)
	if d.height >= 512 {
		hiPad = 3 + 3 // 3 mode bits + 3 low bits
	} else if d.height >= 256 {
		hiPad = 3 + 4 // 3 mode bits + 4 low bits
	}

	// start transfer
	d.csPin.High()

	for i := int16(0); i < d.height; i++ {
		if d.diffing {
			// Skip rendering lines that haven't changed.
			linediv := (i + 1) / 8
			linemod := uint8((i + 1) % 8)
			if !hasBit(d.lineDiff[linediv], linemod) {
				continue
			}
		}

		// The first 5 bits are either dummy or part of the current line
		// (1-indexed) if it overflows 8-bits.
		// The last 3 bits are the command for the first line and dummy bits
		// for subsequent lines (set as command for simplicity)
		hi := uint8((i + 1) >> 8)
		hi = hi << hiPad
		d.txBuf[0] = cmd | hi

		// The second byte is the low bits of the current line (1-indexed).
		// for <8 bits cases, the high bits are dummy, so we leave them as 0.
		d.txBuf[1] = uint8(i + 1)

		// send the first two bytes
		err := d.bus.Tx(d.txBuf, nil)
		if err != nil {
			return err
		}

		// send the line data
		err = d.bus.Tx(d.buffer[i*d.bytesPerLine:(i+1)*d.bytesPerLine], nil)
		if err != nil {
			return err
		}
	}

	// Trailer 16 bits (low)
	d.txBuf[0] = 0x00
	d.txBuf[1] = 0x00
	err := d.bus.Tx(d.txBuf, nil)
	if err != nil {
		return err
	}

	// end transfer
	d.csPin.Low()

	return nil
}

// holdDisplay simply toggles VCOM without updating any lines.
func (d *Device) holdDisplay() error {
	d.txBuf[0] = d.vcom
	d.txBuf[1] = 0x00

	d.toggleVcom()

	// begin transaction
	d.csPin.High()

	err := d.bus.Tx(d.txBuf, nil)
	if err != nil {
		return err
	}

	// end transaction
	d.csPin.Low()

	return nil
}

// Clear clears both the in-memory buffer and the display.
func (d *Device) Clear() error {
	if d.width == 0 {
		return errors.New("display not configured")
	}

	d.ClearBuffer()
	return d.ClearDisplay()
}

// ClearBuffer clears the in-memory buffer. The display is not updated.
func (d *Device) ClearBuffer() {
	if d.width == 0 {
		return
	}

	if d.diffing {
		// detect what rows need to be reset on the next render
		d.invalidateModifiedLines()
	}

	// reset the in-memory buffer
	for i := 0; i < len(d.buffer); i++ {
		d.buffer[i] = 0xff
	}
}

// invalidateModifiedLines marks any line that has at least a single black pixel
// as invalidated. Padding bits, if any, are always 1.
func (d *Device) invalidateModifiedLines() {
	for y := int16(0); y < d.height; y++ {
		offset := y * d.bytesPerLine

		updateLine := false
		for x := int16(0); x < d.width; x++ {
			div := offset + x/8
			mod := uint8(x % 8)

			if !hasBit(d.buffer[div], mod) {
				updateLine = true
				break
			}
		}

		if updateLine {
			d.invalidateLine(y)
		}
	}
}

// ClearDisplay clears the display. The in-memory buffer is not updated. A
// subsequent call to Display() will re-render the content as it was before
// clearing.
func (d *Device) ClearDisplay() error {
	if d.width == 0 {
		return errors.New("display not configured")
	}

	d.txBuf[0] = d.vcom | bitClear
	d.txBuf[1] = 0x00

	d.toggleVcom()

	// begin transaction
	d.csPin.High()

	err := d.bus.Tx(d.txBuf, nil)
	if err != nil {
		return err
	}

	// end transaction
	d.csPin.Low()

	return nil
}

// invalidateLine marks a line and the frame itself as invalidated.
func (d *Device) invalidateLine(line int16) {
	// mark the frame as invalidated
	d.lineDiff[0] = setBit(d.lineDiff[0], 0)

	// mark the line as invalidated
	linediv := (line + 1) / 8
	linemod := uint8((line + 1) % 8)
	d.lineDiff[linediv] = setBit(d.lineDiff[linediv], linemod)
}

// toggleVcom toggles the VCOM, as is instructed by the datasheet.
// Toggling VCOM can help maintain the display's longevity. It should ideally
// be called at least once per second, preferably at 4-100 Hz.
// Toggling VCOM causes a tiny bit of flicker, but without it the pixels can
// be permanently damaged by the DC bias accumulating over time.
func (d *Device) toggleVcom() {
	if d.vcom != 0 {
		d.vcom = 0x00
	} else {
		d.vcom = bitVcom
	}
}
