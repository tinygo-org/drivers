package hd44780

import (
	"errors"
	"image/color"
	"time"

	"machine"
)

type Buser interface {
	Write(data byte)
	Read() byte
	SetCommandMode(set bool)
}

type GPIO struct {
	dataPins []machine.GPIO
	e        machine.GPIO
	rw       machine.GPIO
	rs       machine.GPIO

	write func(data byte)
	read  func() uint8
}

// NewGPIO4Bit returns 4bit data length HD44780 driver
func NewGPIO4Bit(data []uint8, e, rs, rw uint8) (Device, error) {
	const fourBitMode = 4
	if len(data) != fourBitMode {
		return Device{}, errors.New("4 pins are required in data slice (D7-D4) when HD44780 is used in 4 bit mode")
	}
	return newGPIO(data, e, rs, rw, DATA_LENGTH_4BIT), nil
}

// NewGPIO8Bit returns 8bit data length HD44780 driver
func NewGPIO8Bit(data []uint8, e, rs, rw uint8) (Device, error) {
	const eightBitMode = 8
	if len(data) != eightBitMode {
		return Device{}, errors.New("8 pins are required in data slice (D7-D0) when HD44780 is used in 8 bit mode")
	}
	return newGPIO(data, e, rs, rw, DATA_LENGTH_8BIT), nil
}

func newGPIO(data []uint8, e, rs, rw uint8, mode byte) Device {
	pins := make([]machine.GPIO, len(data))
	for i := 0; i < len(data); i++ {
		m := machine.GPIO{Pin: data[i]}
		m.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
		pins[i] = m
	}
	enable := machine.GPIO{e}
	enable.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	registerSelect := machine.GPIO{rs}
	registerSelect.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	readWrite := machine.GPIO{rw}
	readWrite.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	readWrite.Low()

	gpio := GPIO{
		dataPins: pins,
		e:        enable,
		rs:       registerSelect,
		rw:       readWrite,
	}

	if mode == DATA_LENGTH_4BIT {
		gpio.write = gpio.write4BitMode
		gpio.read = gpio.read4BitMode
	} else {
		gpio.write = gpio.write8BitMode
		gpio.read = gpio.read8BitMode
	}

	return Device{
		bus:        &gpio,
		datalength: mode,
	}
}

// SetCommandMode sets command/instruction mode
func (g *GPIO) SetCommandMode(set bool) {
	if set {
		g.rs.Low()
	} else {
		g.rs.High()
	}
}

func (g *GPIO) Write(data byte) {
	g.rw.Low()
	g.write(data)
}

func (g *GPIO) write8BitMode(data byte) {
	g.e.High()
	g.setPins(data)
	g.e.Low()
}

func (g *GPIO) write4BitMode(data byte) {
	g.e.High()
	g.setPins(data >> 4)
	g.e.Low()

	g.e.High()
	g.setPins(data)
	g.e.Low()
}

func (g *GPIO) Read() byte {
	g.rs.Low()
	g.rw.High()
	g.reconfigureGPIOMode(machine.GPIO_INPUT)
	data := g.read()
	g.reconfigureGPIOMode(machine.GPIO_OUTPUT)
	return data
}

func (g *GPIO) read4BitMode() byte {
	g.e.High()
	data := (g.pins() << 4 & 0xF0)
	g.e.Low()
	g.e.High()
	data |= (g.pins() & 0x0F)
	g.e.Low()
	return data
}
func (g *GPIO) read8BitMode() byte {
	g.e.High()
	data := g.pins()
	g.e.Low()
	return data
}
func (g *GPIO) reconfigureGPIOMode(mode machine.GPIOMode) {
	for i := 0; i < len(g.dataPins); i++ {
		g.dataPins[i].Configure(machine.GPIOConfig{Mode: mode})
	}
}

// setPins sets high or low state on all data pins depending on data
func (g *GPIO) setPins(data uint8) {
	mask := uint8(1)
	for i := 0; i < len(g.dataPins); i++ {
		if (data & mask) != 0 {
			g.dataPins[i].High()
		} else {
			g.dataPins[i].Low()
		}
		mask = mask << 1
	}
}

// pins returns current state of data pins. MSB is D7
func (g *GPIO) pins() uint8 {
	bits := uint8(0)
	for i := uint8(0); i < uint8(len(g.dataPins)); i++ {
		if g.dataPins[i].Get() {
			bits |= (1 << i)
		}
	}
	return bits
}

type Device struct {
	bus        Buser
	width      uint8
	height     uint8
	buffer     []uint8
	bufferSize uint8

	rowOffset  []uint8 // Row offsets in DDRAM
	datalength uint8

	cursor cursor
}

type cursor struct {
	x, y uint8
}

type Config struct {
	Width       int16
	Height      int16
	CursorBlink bool
	CursorOnOff bool
	Font        uint8
}

// Configure initializes device
func (d *Device) Configure(cfg Config) error {
	d.width = uint8(cfg.Width)
	d.height = uint8(cfg.Height)
	if d.width == 0 || d.height == 0 {
		return errors.New("Width and height must be set")
	}
	memoryMap := uint8(ONE_LINE)
	if d.height > 1 {
		memoryMap = TWO_LINE
	}
	d.setRowOffsets()
	d.ClearBuffer()

	cursor := CURSOR_OFF
	if cfg.CursorOnOff {
		cursor = CURSOR_ON
	}
	cursorBlink := CURSOR_BLINK_OFF
	if cfg.CursorBlink {
		cursorBlink = CURSOR_BLINK_ON
	}
	if !(cfg.Font == FONT_5X8 || cfg.Font == FONT_5X10) {
		cfg.Font = FONT_5X8
	}

	//Wait 15ms after Vcc rises to 4.5V
	time.Sleep(15 * time.Millisecond)

	d.bus.SetCommandMode(true)
	d.bus.Write(DATA_LENGTH_8BIT)
	time.Sleep(5 * time.Millisecond)

	for i := 0; i < 2; i++ {
		d.bus.Write(DATA_LENGTH_8BIT)
		time.Sleep(150 * time.Microsecond)
	}

	if d.datalength == DATA_LENGTH_4BIT {
		d.bus.Write(DATA_LENGTH_4BIT >> 4)
	}

	// Busy flag is now accessible
	d.SendCommand(memoryMap | cfg.Font | d.datalength)
	d.SendCommand(DISPLAY_OFF)
	d.SendCommand(DISPLAY_CLEAR)
	d.SendCommand(ENTRY_MODE | CURSOR_INCREASE | DISPLAY_NO_SHIFT)
	d.SendCommand(DISPLAY_ON | uint8(cursor) | uint8(cursorBlink))
	return nil
}

// Write writes data to internal buffer
func (d *Device) Write(data []byte) (n int, err error) {
	size := len(data)
	if size > len(d.buffer) {
		size = len(d.buffer)
	}
	d.bufferSize = uint8(size)
	for i := uint8(0); i < d.bufferSize; i++ {
		d.buffer[i] = data[i]
	}
	return size, nil
}

// Display sends the whole buffer to the screen at cursor position
func (d *Device) Display() error {
	var totalDisplayed uint8
	var bufferX uint8
	var bufferY uint8
	var curPosX uint8

	for ; d.cursor.y < d.height; d.cursor.y++ {
		d.SetCursor(d.cursor.x, d.cursor.y)

		for curPosX = d.cursor.x; curPosX < d.width && totalDisplayed < d.bufferSize; curPosX++ {
			d.SendData(d.buffer[bufferY*(d.width)+bufferX])
			bufferX++
			totalDisplayed++
		}
		if bufferX > d.width {
			bufferX = 0
			bufferY++
		}
		if totalDisplayed >= d.bufferSize {
			d.cursor.x = curPosX
			break
		}
		if curPosX >= d.width {
			curPosX = 0
		}
		d.cursor.x = curPosX
	}
	return nil
}

// SetCursor moves cursor to position x,y, where (0,0) is top left corner and (width-1, height-1) bottom right
func (d *Device) SetCursor(x, y uint8) {
	d.cursor.x = x
	d.cursor.y = y
	d.SendCommand(DDRAM_SET | (x + (d.rowOffset[y] * y)))
}

// SetRowOffsets sets initial memory addresses coresponding to the display rows
// Each row on display has different starting address in DDRAM. Rows are not mapped in order.
// These addresses tend to differ between the types of the displays (16x2, 16x4, 20x4 etc ..),
// https://web.archive.org/web/20111122175541/http://web.alfredstate.edu/weimandn/lcd/lcd_addressing/lcd_addressing_index.html
func (d *Device) setRowOffsets() {
	switch d.height {
	case 1:
		d.rowOffset = []uint8{}
	case 2:
		d.rowOffset = []uint8{0x0, 0x40, 0x0, 0x40}
	case 4:
		d.rowOffset = []uint8{0x0, 0x40, d.width, 0x40 + d.width}
	default:
		d.rowOffset = []uint8{0x0, 0x40, d.width, 0x40 + d.width}

	}
}

// SendCommand sends commands to driver
func (d *Device) SendCommand(command uint8) {
	d.bus.SetCommandMode(true)
	d.bus.Write(command)

	for d.Busy() {
	}
}

// SendData sends byte data directly to display.
func (d *Device) SendData(data uint8) {
	d.bus.SetCommandMode(false)
	d.bus.Write(data)

	for d.Busy() {
	}
}

// CreateCharacter crates characters using data and stores it under cgram Addr in CGRAM
func (d *Device) CreateCharacter(cgramAddr uint8, data []byte) {
	d.SendCommand(CGRAM_SET | cgramAddr)
	for _, dd := range data {
		d.SendData(dd)
	}
}

// Busy returns true when hd447890 is busy
func (d *Device) Busy() bool {
	status := d.bus.Read()
	return (status & BUSY) > 0
}

// SetPixel is not supported on devices which uses HD44780 driver
func (d *Device) SetPixel(x, y int16, c color.RGBA) {
	panic("HD44780 does not support setting individual pixels")
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return int16(d.width), int16(d.height)
}

// ClearDisplay clears displayed content and buffer
func (d *Device) ClearDisplay() {
	d.SendCommand(DISPLAY_CLEAR)
	d.ClearBuffer()
}

// ClearBuffer clears internal buffer
func (d *Device) ClearBuffer() {
	d.buffer = make([]uint8, d.width*d.height)
}
