// Package adafruit4650 implements a driver for the Adafruit FeatherWing OLED - 128x64 OLED display.
// The display is backed itself by a SH1107 driver chip.
//
// Store: https://www.adafruit.com/product/4650
//
// Documentation: https://learn.adafruit.com/adafruit-128x64-oled-featherwing
package adafruit4650

import (
	"image/color"
	"time"

	"tinygo.org/x/drivers"
)

const DefaultAddress = 0x3c

const (
	commandSetLowColumn  = 0x00
	commandSetHighColumn = 0x10
	commandSetPage       = 0xb0
)

const (
	width  = 128
	height = 64
)

// Device represents an Adafruit 4650 device
type Device struct {
	bus     drivers.I2C
	Address uint8
	buffer  []byte
	width   int16
	height  int16
}

// New creates a new device, not configuring anything yet.
func New(bus drivers.I2C) Device {
	return Device{
		bus:     bus,
		Address: DefaultAddress,
		width:   width,
		height:  height,
	}
}

// Configure initializes the display with default configuration
func (d *Device) Configure() error {

	bufferSize := d.width * d.height / 8
	d.buffer = make([]byte, bufferSize)

	// This sequence is an amalgamation of the datasheet, official Arduino driver, CircuitPython driver and other drivers
	initSequence := []byte{
		0xae, // display off, sleep mode
		//0xd5, 0x41, // set display clock divider (from original datasheet)
		0xd5, 0x51, // set display clock divider (from Adafruit driver)
		0xd9, 0x22, // pre-charge/dis-charge period mode: 2 DCLKs/2 DCLKs (POR)
		0x20,       // memory mode
		0x81, 0x4f, // contrast setting = 0x4f
		0xad, 0x8a, // set dc/dc pump
		0xa0,       // segment remap, flip-x
		0xc0,       // common output scan direction
		0xdc, 0x00, // set display start line 0 (POR=0)
		0xa8, 0x3f, // multiplex ratio, height - 1 = 0x3f
		0xd3, 0x60, // set display offset mode = 0x60
		0xdb, 0x35, // VCOM deselect level = 0.770 (POR)
		0xa4, // entire display off, retain RAM, normal status (POR)
		0xa6, // normal (not reversed) display
		0xaf, // display on
	}

	err := d.writeCommands(initSequence)
	if err != nil {
		return err
	}

	// recommended in the datasheet, same in other drivers
	time.Sleep(100 * time.Millisecond)

	return nil
}

// ClearDisplay clears the image buffer as well as the actual display
func (d *Device) ClearDisplay() error {
	d.ClearBuffer()
	return d.Display()
}

// ClearBuffer clears the buffer
func (d *Device) ClearBuffer() {
	bzero(d.buffer)
}

// SetPixel modifies the internal buffer. Since this display has a bit-depth of 1 bit any non-zero
// color component will be treated as 'on',  otherwise 'off'.
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return
	}

	// RAM layout
	//    *-----> y
	//    |
	//   x|     col0  col1  ... col63
	//    v  p0  a0    b0         ..
	//           a1    b1         ..
	//           ..    ..         ..
	//           a7    b7         ..
	//       p1  a0    b0
	//           a1    b1
	//

	//flip y - so the display orientation matches the silk screen labeling etc.
	y = d.height - y - 1

	page := x / 8
	bytesPerPage := d.height
	byteIndex := y + bytesPerPage*page
	bit := x % 8
	if (c.R | c.G | c.B) != 0 {
		d.buffer[byteIndex] |= 1 << uint8(bit)
	} else {
		d.buffer[byteIndex] &^= 1 << uint8(bit)
	}
}

// Display sends the whole buffer to the screen
func (d *Device) Display() error {

	bytesPerPage := d.height

	pages := (d.width + 7) / 8
	for page := int16(0); page < pages; page++ {

		err := d.setRAMPosition(uint8(page), 0)
		if err != nil {
			return err
		}

		offset := page * bytesPerPage
		err = d.writeRAM(d.buffer[offset : offset+bytesPerPage])
		if err != nil {
			return err
		}
	}

	return nil
}

// setRAMPosition updates the device's current page and column position
func (d *Device) setRAMPosition(page uint8, column uint8) error {
	if page > 15 {
		panic("page out of bounds")
	}
	if column > 127 {
		panic("column out of bounds")
	}
	setPage := commandSetPage | (page & 0xF)

	lo := column & 0xF
	setLowColumn := commandSetLowColumn | lo

	hi := (column >> 4) & 0x7
	setHighColumn := commandSetHighColumn | hi

	cmds := []byte{
		setPage,
		setLowColumn,
		setHighColumn,
	}

	return d.writeCommands(cmds)
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return d.width, d.height
}

func (d *Device) writeCommands(commands []byte) error {
	onlyCommandsFollowing := byte(0x00)
	return d.bus.Tx(uint16(d.Address), append([]byte{onlyCommandsFollowing}, commands...), nil)
}

func (d *Device) writeRAM(data []byte) error {
	onlyRAMFollowing := byte(0x40)
	return d.bus.Tx(uint16(d.Address), append([]byte{onlyRAMFollowing}, data...), nil)
}

func bzero(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}
}
