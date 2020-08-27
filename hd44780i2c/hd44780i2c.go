// Package hd44780i2c implements a driver for the Hitachi HD44780 LCD display module
// with an I2C adapter.
//
// Datasheet: https://www.sparkfun.com/datasheets/LCD/HD44780.pdf
//
package hd44780i2c

import (
	"errors"
	"time"

	"tinygo.org/x/drivers"
)

// Device wraps an I2C connection to a HD44780 I2C LCD with related data.
type Device struct {
	bus             drivers.I2C
	addr            uint8
	width           uint8
	height          uint8
	cursor          cursor
	backlight       uint8
	displayfunction uint8
	displaycontrol  uint8
	displaymode     uint8
}

type cursor struct {
	x, y uint8
}

// Config for HD44780 I2C LCD.
type Config struct {
	Width       uint8
	Height      uint8
	Font        uint8
	CursorOn    bool
	CursorBlink bool
}

// New creates a new HD44780 I2C LCD connection. The I2C bus must already be
// configured.
//
// This function only creates the Device object, it does not touch the device.
func New(bus drivers.I2C, addr uint8) Device {
	if addr == 0 {
		addr = 0x27
	}
	return Device{
		bus:  bus,
		addr: addr,
	}
}

// Configure sets up the display. Display itself and backlight is default on.
func (d *Device) Configure(cfg Config) error {

	if cfg.Width == 0 || cfg.Height == 0 {
		return errors.New("width and height must be set")
	}
	d.width = uint8(cfg.Width)
	d.height = uint8(cfg.Height)

	delayms(50)

	d.backlight = BACKLIGHT_ON
	d.expanderWrite(0)
	delayms(1000)

	d.write4bits(0x03 << 4)
	delayus(4500)
	d.write4bits(0x03 << 4)
	delayus(4500)
	d.write4bits(0x03 << 4)
	delayus(150)
	d.write4bits(0x02 << 4)

	d.displayfunction = DATA_LENGTH_4BIT | ONE_LINE | FONT_5X8
	if d.height > 1 {
		d.displayfunction |= TWO_LINE
	}
	if cfg.Font != 0 && d.height == 1 {
		d.displayfunction |= FONT_5X10
	}
	d.sendCommand(FUNCTION_MODE | d.displayfunction)

	d.displaycontrol = DISPLAY_ON | CURSOR_OFF | CURSOR_BLINK_OFF
	if cfg.CursorOn {
		d.displaycontrol |= CURSOR_ON
	}
	if cfg.CursorBlink {
		d.displaycontrol |= CURSOR_BLINK_ON
	}
	d.sendCommand(DISPLAY_ON_OFF | d.displaycontrol)
	d.ClearDisplay()

	d.displaymode = CURSOR_INCREASE | DISPLAY_NO_SHIFT
	d.sendCommand(ENTRY_MODE | d.displaymode)
	d.Home()

	return nil
}

// ClearDisplay clears all texts on the display.
func (d *Device) ClearDisplay() {
	d.sendCommand(DISPLAY_CLEAR)
	d.cursor.x = 0
	d.cursor.y = 0
	delayus(2000)
}

// Home sets the cursor back to position (0, 0).
func (d *Device) Home() {
	d.sendCommand(CURSOR_HOME)
	d.cursor.x = 0
	d.cursor.y = 0
	delayus(2000)
}

// SetCursor sets the cursor to a specific position (x, y).
//
// if y (row) is set larger than actual rows, it would be set to 0.
func (d *Device) SetCursor(x, y uint8) {
	rowOffset := []uint8{0x0, 0x40, 0x14, 0x54}
	if y > (d.height - 1) {
		y = 0
	}
	d.cursor.x = x
	d.cursor.y = y
	d.sendCommand(DDRAM_SET | (x + (rowOffset[y])))
}

// Print prints text on the display (started from current cursor position).
//
// It would automatically break to new line when the text is too long.
// You can also use \n as line breakers.
func (d *Device) Print(data []byte) {
	for _, chr := range data {
		if chr == '\n' {
			d.newLine()
		} else {
			d.cursor.x++
			if d.cursor.x >= d.width {
				d.newLine()
			}
			d.sendData(uint8(rune(chr)))
		}
	}
}

// CreateCharacter crates custom characters (using data parameter)
// and stores it under CGRAM address (using cgramAddr, 0x0-0x7).
func (d *Device) CreateCharacter(cgramAddr uint8, data []byte) {
	cgramAddr &= 0x7
	d.sendCommand(CGRAM_SET | cgramAddr<<3)
	for _, dd := range data {
		d.sendData(dd)
	}
	d.SetCursor(d.cursor.x, d.cursor.y)
}

// DisplayOn turns on/off the display.
func (d *Device) DisplayOn(option bool) {
	if option {
		d.displaycontrol |= DISPLAY_ON
	} else {
		d.displaycontrol &= ^uint8(DISPLAY_ON)
	}
	d.sendCommand(DISPLAY_ON_OFF | d.displaycontrol)
}

// CursorOn display/hides the cursor.
func (d *Device) CursorOn(option bool) {
	if option {
		d.displaycontrol |= CURSOR_ON
	} else {
		d.displaycontrol &= ^uint8(CURSOR_ON)
	}
	d.sendCommand(DISPLAY_ON_OFF | d.displaycontrol)
}

// CursorBlink turns on/off the blinking cursor mode.
func (d *Device) CursorBlink(option bool) {
	if option {
		d.displaycontrol |= CURSOR_BLINK_ON
	} else {
		d.displaycontrol &= ^uint8(CURSOR_BLINK_ON)
	}
	d.sendCommand(DISPLAY_ON_OFF | d.displaycontrol)
}

// BacklightOn turns on/off the display backlight.
func (d *Device) BacklightOn(option bool) {
	if option {
		d.backlight = BACKLIGHT_ON
	} else {
		d.backlight = BACKLIGHT_OFF
	}
	d.expanderWrite(0)
}

func (d *Device) newLine() {
	d.cursor.x = 0
	d.cursor.y++
	d.SetCursor(d.cursor.x, d.cursor.y)
}

func delayms(t uint16) {
	time.Sleep(time.Millisecond * time.Duration(t))
}

func delayus(t uint16) {
	time.Sleep(time.Microsecond * time.Duration(t))
}

func (d *Device) expanderWrite(value uint8) {
	d.bus.Tx(uint16(d.addr), []uint8{value | d.backlight}, nil)
}

func (d *Device) pulseEnable(value uint8) {
	d.expanderWrite(value | En)
	delayus(1)
	d.expanderWrite(value & ^uint8(En))
	delayus(50)
}

func (d *Device) write4bits(value uint8) {
	d.expanderWrite(value)
	d.pulseEnable(value)
}

func (d *Device) write(value uint8, mode uint8) {
	d.write4bits(uint8(value&0xf0) | mode)
	d.write4bits(uint8((value<<4)&0xf0) | mode)
}

func (d *Device) sendCommand(value uint8) {
	d.write(value, 0)
}

func (d *Device) sendData(value uint8) {
	d.write(value, Rs)
}
