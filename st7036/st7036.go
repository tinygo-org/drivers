// Package st7036 provides a driver for the Sitronix ST7036 character LCD.
// It supports the Pimoroni Display-o-Tron HAT and Display-o-Tron 3000.
//
// Each byte uses a separate chip-select pulse on the 3-wire SPI bus.
// See https://www.crystalfontz.com/controllers/Sitronix/ST7036/
//
// The command sequence follows https://github.com/pimoroni/st7036
package st7036 // import "tinygo.org/x/drivers/st7036"

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

var (
	errInvalidContrast = errors.New("st7036: contrast must be in the range 0..0x3F")
	errInvalidPosition = errors.New("st7036: row and column must be within the display size")
	errInvalidCharSlot = errors.New("st7036: character slot must be in the range 0..7")
)

// Config holds the display size. Both Pimoroni boards have 16 columns and 3 rows.
type Config struct {
	Rows    uint8 // number of rows, 1 to 3. Zero defaults to 3.
	Columns uint8 // number of columns. Zero defaults to 16.
}

// Device is a handle to a ST7036 character LCD.
type Device struct {
	bus   drivers.SPI
	cs    outputPin
	rs    outputPin
	reset outputPin

	rows       uint8
	columns    uint8
	rowOffsets []uint8

	instructionSetTemplate uint8
	doubleHeight           uint8

	enabled       bool
	cursorEnabled bool
	cursorBlink   bool
}

type outputPin interface {
	High()
	Low()
}

// Configure resets and clears the display with 1/5 bias and a contrast of 40.
func (d *Device) Configure(cfg Config) error {
	d.rows = cfg.Rows
	if d.rows == 0 {
		d.rows = 3
	}
	d.columns = cfg.Columns
	if d.columns == 0 {
		d.columns = 16
	}
	switch d.rows {
	case 1:
		d.rowOffsets = []uint8{0x00}
	case 2:
		d.rowOffsets = []uint8{0x00, 0x40}
	case 3:
		d.rowOffsets = []uint8{0x00, 0x10, 0x20}
	default:
		return errInvalidPosition
	}

	d.Reset()

	d.enabled = true
	d.updateDisplayMode()

	d.writeCommand(entryMove|entryShift, 0)

	d.SetBias(1)

	if err := d.SetContrast(40); err != nil {
		return err
	}

	return d.Clear()
}

// Reset pulses the reset line, if connected. It is a no-op otherwise.
func (d *Device) Reset() {
	if d.reset == nil {
		return
	}
	d.reset.Low()
	time.Sleep(time.Millisecond)
	d.reset.High()
	time.Sleep(time.Millisecond)
}

// Size returns the number of columns and rows on the display.
func (d *Device) Size() (columns, rows uint8) {
	return d.columns, d.rows
}

// Clear clears the display and homes the cursor.
func (d *Device) Clear() error {
	d.writeCommand(commandClear, 0)
	time.Sleep(1500 * time.Microsecond)
	return d.Home()
}

// Home moves the cursor to column 0, row 0.
func (d *Device) Home() error {
	return d.SetCursorPosition(0, 0)
}

// SetContrast sets the display contrast, in the range 0 to 0x3F.
func (d *Device) SetContrast(contrast uint8) error {
	if contrast > 0x3F {
		return errInvalidContrast
	}

	d.writeCommand(0b01010100|((contrast>>4)&0x03), 1)
	d.writeCommand(0b01101011, 1)

	d.writeCommand(0b01110000|(contrast&0x0F), 1)
	return nil
}

// SetDisplayMode enables or disables the display, cursor and cursor blink.
func (d *Device) SetDisplayMode(enable, cursor, blink bool) {
	d.enabled = enable
	d.cursorEnabled = cursor
	d.cursorBlink = blink
	d.updateDisplayMode()
}

func (d *Device) updateDisplayMode() {
	mask := uint8(commandSetDisplayMode)
	if d.enabled {
		mask |= displayOn
	}
	if d.cursorEnabled {
		mask |= cursorOn
	}
	if d.cursorBlink {
		mask |= blinkOn
	}
	d.writeCommand(mask, 0)
}

// SetBias sets the LCD bias. Use 1 for 1/5 bias on a 3-row display.
func (d *Device) SetBias(bias uint8) {
	d.writeCommand(commandBias|(bias<<4)|1, 1)
}

// SetCursorOffset sets the cursor position directly, as a DDRAM address.
func (d *Device) SetCursorOffset(offset uint8) {
	d.writeCommand(setDDRAM|offset, 0)
}

// SetCursorPosition moves the cursor to the given column and row.
func (d *Device) SetCursorPosition(column, row uint8) error {
	if row >= d.rows || column >= d.columns {
		return errInvalidPosition
	}
	d.SetCursorOffset(d.rowOffsets[row] + column)
	time.Sleep(1500 * time.Microsecond)
	return nil
}

// CreateChar defines a character in slot 0 to 7.
// Each byte sets one pixel row. The lower 5 bits set the pixels.
func (d *Device) CreateChar(slot uint8, charMap [8]byte) error {
	if slot > 7 {
		return errInvalidCharSlot
	}
	base := slot * 8
	for i := uint8(0); i < 8; i++ {
		d.writeCommand(setCGRAM|(base+i), 0)
		d.writeChar(charMap[i])
	}
	return d.Home()
}

// Write sends bytes at the current cursor position and implements io.Writer.
// Bytes 0 to 7 select custom characters.
func (d *Device) Write(data []byte) (int, error) {
	d.rs.High()
	for _, b := range data {
		d.transfer(b)
		time.Sleep(50 * time.Microsecond)
	}
	return len(data), nil
}

func (d *Device) writeChar(value byte) {
	d.rs.High()
	d.transfer(value)
	time.Sleep(100 * time.Microsecond)
}

// writeCommand selects the given instruction set, then sends value as a
// command byte.
func (d *Device) writeCommand(value, instructionSet uint8) {
	d.rs.Low()
	d.writeInstructionSet(instructionSet)
	d.transfer(value)
	time.Sleep(60 * time.Microsecond)
}

func (d *Device) writeInstructionSet(instructionSet uint8) {
	d.rs.Low()
	d.transfer(d.instructionSetTemplate | instructionSet | (d.doubleHeight << 2))
	time.Sleep(60 * time.Microsecond)
}

func (d *Device) transfer(b byte) {
	d.cs.Low()
	d.bus.Transfer(b)
	d.cs.High()
}
