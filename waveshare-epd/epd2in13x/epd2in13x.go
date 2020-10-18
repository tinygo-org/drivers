// Package epd2in13x implements a driver for Waveshare 2.13in (B & C versions) tri-color e-paper device.
//
// Datasheet: https://www.waveshare.com/w/upload/d/d3/2.13inch-e-paper-b-Specification.pdf
//
package epd2in13x // import "tinygo.org/x/drivers/waveshare-epd/epd2in13x"

import (
	"errors"
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

type Config struct {
	Width     int16
	Height    int16
	NumColors uint8
}

type Device struct {
	bus          drivers.SPI
	cs           machine.Pin
	dc           machine.Pin
	rst          machine.Pin
	busy         machine.Pin
	width        int16
	height       int16
	buffer       [][]uint8
	bufferLength uint32
}

type Color uint8

// New returns a new epd2in13x driver. Pass in a fully configured SPI bus.
func New(bus drivers.SPI, csPin, dcPin, rstPin, busyPin machine.Pin) Device {
	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	dcPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rstPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	busyPin.Configure(machine.PinConfig{Mode: machine.PinInput})
	return Device{
		bus:  bus,
		cs:   csPin,
		dc:   dcPin,
		rst:  rstPin,
		busy: busyPin,
	}
}

// Configure sets up the device.
func (d *Device) Configure(cfg Config) {
	if cfg.Width != 0 {
		d.width = cfg.Width
	} else {
		d.width = 104
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	} else {
		d.height = 212
	}
	if cfg.NumColors == 0 {
		cfg.NumColors = 3
	} else if cfg.NumColors == 1 {
		cfg.NumColors = 2
	}
	d.bufferLength = (uint32(d.width) * uint32(d.height)) / 8
	d.buffer = make([][]uint8, cfg.NumColors-1)
	for i := range d.buffer {
		d.buffer[i] = make([]uint8, d.bufferLength)
	}
	for i := range d.buffer {
		for j := uint32(0); j < d.bufferLength; j++ {
			d.buffer[i][j] = 0xFF
		}
	}

	d.cs.Low()
	d.dc.Low()
	d.rst.Low()

	d.Reset()

	d.SendCommand(BOOSTER_SOFT_START)
	d.SendData(0x17)
	d.SendData(0x17)
	d.SendData(0x17)
	d.SendCommand(POWER_ON)
	d.WaitUntilIdle()
	d.SendCommand(PANEL_SETTING)
	d.SendData(0x8F)
	d.SendCommand(VCOM_AND_DATA_INTERVAL_SETTING)
	d.SendData(0x37)
	d.SendCommand(RESOLUTION_SETTING)
	d.SendData(uint8(d.width))
	d.SendData(0x00)
	d.SendData(uint8(d.height))
}

// Reset resets the device
func (d *Device) Reset() {
	d.rst.Low()
	time.Sleep(200 * time.Millisecond)
	d.rst.High()
	time.Sleep(200 * time.Millisecond)
}

// DeepSleep puts the display into deepsleep
func (d *Device) DeepSleep() {
	d.SendCommand(POWER_OFF)
	d.WaitUntilIdle()
	d.SendCommand(DEEP_SLEEP)
	d.SendData(0xA5)
}

// SendCommand sends a command to the display
func (d *Device) SendCommand(command uint8) {
	d.sendDataCommand(true, command)
}

// SendData sends a data byte to the display
func (d *Device) SendData(data uint8) {
	d.sendDataCommand(false, data)
}

// sendDataCommand sends image data or a command to the screen
func (d *Device) sendDataCommand(isCommand bool, data uint8) {
	if isCommand {
		d.dc.Low()
	} else {
		d.dc.High()
	}
	d.cs.Low()
	d.bus.Transfer(data)
	d.cs.High()
}

// SetPixel modifies the internal buffer in a single pixel.
// The display have 3 colors: black, white and a third color that could be red or yellow
// We use RGBA(0,0,0, 255) as white (transparent)
// RGBA(1-255,0,0,255) as colored (red or yellow)
// Anything else as black
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return
	}
	if c.R != 0 && c.G == 0 && c.B == 0 { // COLORED
		d.SetEPDPixel(x, y, COLORED)
	} else if c.G != 0 || c.B != 0 { // BLACK
		d.SetEPDPixel(x, y, BLACK)
	} else { // WHITE / EMPTY
		d.SetEPDPixel(x, y, WHITE)
	}
}

// SetEPDPixel modifies the internal buffer in a single pixel.
func (d *Device) SetEPDPixel(x int16, y int16, c Color) {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return
	}
	byteIndex := (x + y*d.width) / 8
	if c == WHITE {
		d.buffer[BLACK-1][byteIndex] |= 0x80 >> uint8(x%8)
		d.buffer[COLORED-1][byteIndex] |= 0x80 >> uint8(x%8)
	} else if c == COLORED {
		d.buffer[BLACK-1][byteIndex] |= 0x80 >> uint8(x%8)
		d.buffer[COLORED-1][byteIndex] &^= 0x80 >> uint8(x%8)
	} else { // BLACK
		d.buffer[COLORED-1][byteIndex] |= 0x80 >> uint8(x%8)
		d.buffer[BLACK-1][byteIndex] &^= 0x80 >> uint8(x%8)
	}
}

// Display sends the buffer (if any) to the screen.
func (d *Device) Display() error {
	d.SendCommand(DATA_START_TRANSMISSION_1) // black
	time.Sleep(2 * time.Millisecond)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.SendData(d.buffer[BLACK-1][i])
	}
	time.Sleep(2 * time.Millisecond)
	d.SendCommand(DATA_START_TRANSMISSION_2) // red
	time.Sleep(2 * time.Millisecond)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.SendData(d.buffer[COLORED-1][i])
	}
	time.Sleep(2 * time.Millisecond)
	d.SendCommand(DISPLAY_REFRESH)
	return nil
}

// SetDisplayRect sends a rectangle of data at specific coordinates to the device SRAM directly
func (d *Device) SetDisplayRect(buffer [][]uint8, x int16, y int16, w int16, h int16) error {
	if w%8 != 0 {
		return errors.New("rectangle width needs to be a multiple of 8")
	}
	for i := range buffer {
		if int16(len(buffer[i])) < (w/8)*h {
			return errors.New("buffer has the wrong size")
		}
	}
	d.SendCommand(PARTIAL_IN)
	d.SendCommand(PARTIAL_WINDOW)
	d.SendData(uint8(x) & 0xF8)
	d.SendData(((uint8(x) & 0xF8) + uint8(w) - 1) | 0x07)
	d.SendData(uint8(y) >> 8)
	d.SendData(uint8(y) & 0xFF)
	d.SendData(uint8(y+h-1) >> 8)
	d.SendData(uint8(y+h-1) & 0xFF)
	d.SendData(0x01)
	time.Sleep(2 * time.Millisecond)
	d.SendCommand(DATA_START_TRANSMISSION_1)
	for i := int16(0); i < (w/8)*h; i++ {
		d.SendData(buffer[BLACK-1][i])
	}
	time.Sleep(2 * time.Millisecond)
	if len(buffer) > 1 {
		d.SendCommand(DATA_START_TRANSMISSION_2)
		for i := int16(0); i < (w/8)*h; i++ {
			d.SendData(buffer[COLORED-1][i])
		}
		time.Sleep(2 * time.Millisecond)
	}
	d.SendCommand(PARTIAL_OUT)
	return nil
}

// SetDisplayRectColor sends a rectangle of data at specific coordinates to the device SRAM directly
func (d *Device) SetDisplayRectColor(buffer []uint8, x int16, y int16, w int16, h int16, c Color) error {
	if w%8 != 0 {
		return errors.New("rectangle width needs to be a multiple of 8")
	}
	if int16(len(buffer)) < (w/8)*h {
		return errors.New("buffer has the wrong size")
	}
	if c == WHITE {
		return errors.New("wrong color")
	}
	d.SendCommand(PARTIAL_IN)
	d.SendCommand(PARTIAL_WINDOW)
	d.SendData(uint8(x) & 0xF8)
	d.SendData(((uint8(x) & 0xF8) + uint8(w) - 1) | 0x07)
	d.SendData(uint8(y) >> 8)
	d.SendData(uint8(y) & 0xFF)
	d.SendData(uint8(y+h-1) >> 8)
	d.SendData(uint8(y+h-1) & 0xFF)
	d.SendData(0x01)
	time.Sleep(2 * time.Millisecond)
	if c == COLORED {
		d.SendCommand(DATA_START_TRANSMISSION_2)
	} else {
		d.SendCommand(DATA_START_TRANSMISSION_1)
	}
	for i := int16(0); i < (w/8)*h; i++ {
		d.SendData(buffer[i])
	}
	time.Sleep(2 * time.Millisecond)
	d.SendCommand(PARTIAL_OUT)
	return nil
}

// ClearDisplay erases the device SRAM
func (d *Device) ClearDisplay() {
	d.SendCommand(DATA_START_TRANSMISSION_1) // black
	time.Sleep(2 * time.Millisecond)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.SendData(0xFF)
	}
	time.Sleep(2 * time.Millisecond)
	d.SendCommand(DATA_START_TRANSMISSION_2) // red
	time.Sleep(2 * time.Millisecond)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.SendData(0xFF)
	}
	time.Sleep(2 * time.Millisecond)
}

// WaitUntilIdle waits until the display is ready
func (d *Device) WaitUntilIdle() {
	for !d.busy.Get() {
		time.Sleep(100 * time.Millisecond)
	}
}

// IsBusy returns the busy status of the display
func (d *Device) IsBusy() bool {
	return d.busy.Get()
}

// ClearBuffer sets the buffer to 0xFF (white)
func (d *Device) ClearBuffer() {
	for i := uint8(0); i < uint8(len(d.buffer)); i++ {
		for j := uint32(0); j < d.bufferLength; j++ {
			d.buffer[i][j] = 0xFF
		}
	}
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return d.width, d.height
}
