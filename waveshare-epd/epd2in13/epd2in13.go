// Package epd2in13 implements a driver for Waveshare 2.13in black and white e-paper device.
//
// Datasheet: https://www.waveshare.com/wiki/File:2.13inch_e-Paper_Datasheet.pdf
package epd2in13

import (
	"errors"
	"image/color"
	"machine"
	"time"
)

type Config struct {
	Width  int16
	Height int16
}

type Device struct {
	bus          machine.SPI
	cs           machine.GPIO
	dc           machine.GPIO
	rst          machine.GPIO
	busy         machine.GPIO
	width        int16
	height       int16
	buffer       []uint8
	bufferLength uint32
}

// Look up table for full updates
var lutFullUpdate = [30]uint8{
	0x22, 0x55, 0xAA, 0x55, 0xAA, 0x55, 0xAA, 0x11,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x1E, 0x1E, 0x1E, 0x1E, 0x1E, 0x1E, 0x1E, 0x1E,
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00,
}

// Look up table for partial updates, faster but there will be some ghosting
var lutPartialUpdate = [30]uint8{
	0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x0F, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

// New returns a new epd2in13x driver. Pass in a fully configured SPI bus.
func New(bus machine.SPI, csPin uint8, dcPin uint8, rstPin uint8, busyPin uint8) Device {
	pinCs := machine.GPIO{csPin}
	pinCs.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	pinDc := machine.GPIO{dcPin}
	pinDc.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	pinRst := machine.GPIO{rstPin}
	pinRst.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	pinBusy := machine.GPIO{busyPin}
	pinBusy.Configure(machine.GPIOConfig{Mode: machine.GPIO_INPUT})
	return Device{
		bus:  bus,
		cs:   pinCs,
		dc:   pinDc,
		rst:  pinRst,
		busy: pinBusy,
	}
}

// Configure sets up the device.
func (d *Device) Configure(cfg Config) {
	if cfg.Width != 0 {
		d.width = cfg.Width
	} else {
		d.width = 128
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	} else {
		d.height = 250
	}
	d.bufferLength = (uint32(d.width) * uint32(d.height)) / 8
	d.buffer = make([]uint8, d.bufferLength)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.buffer[i] = 0xFF
	}

	d.cs.Low()
	d.dc.Low()
	d.rst.Low()

	d.Reset()

	d.SendCommand(DRIVER_OUTPUT_CONTROL)
	d.SendData(uint8((d.height - 1) & 0xFF))
	d.SendData(uint8(((d.height - 1) >> 8) & 0xFF))
	d.SendData(0x00) // GD = 0; SM = 0; TB = 0;
	d.SendCommand(BOOSTER_SOFT_START_CONTROL)
	d.SendData(0xD7)
	d.SendData(0xD6)
	d.SendData(0x9D)
	d.SendCommand(WRITE_VCOM_REGISTER)
	d.SendData(0xA8) // VCOM 7C
	d.SendCommand(SET_DUMMY_LINE_PERIOD)
	d.SendData(0x1A) // 4 dummy lines per gate
	d.SendCommand(SET_GATE_TIME)
	d.SendData(0x08) // 2us per line
	d.SendCommand(DATA_ENTRY_MODE_SETTING)
	d.SendData(0x03) // X increment; Y increment

	d.SetLUT(true)
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
	d.SendCommand(DEEP_SLEEP_MODE)
	d.WaitUntilIdle()
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

// SetLUT sets the look up tables for full or partial updates
func (d *Device) SetLUT(fullUpdate bool) {
	d.SendCommand(WRITE_LUT_REGISTER)
	if fullUpdate {
		for i := 0; i < 30; i++ {
			d.SendData(lutFullUpdate[i])
		}
	} else {
		for i := 0; i < 30; i++ {
			d.SendData(lutPartialUpdate[i])
		}
	}
}

// SetPixel modifies the internal buffer in a single pixel.
// The display have 2 colors: black and white
// We use RGBA(0,0,0, 255) as white (transparent)
// Anything else as black
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return
	}
	byteIndex := (x + y*d.width) / 8
	if c.R == 0 && c.G == 0 && c.B == 0 { // TRANSPARENT / WHITE
		d.buffer[byteIndex] |= 0x80 >> uint8(x%8)
	} else { // WHITE / EMPTY
		d.buffer[byteIndex] &^= 0x80 >> uint8(x%8)
	}
}

// Display sends the buffer to the screen.
func (d *Device) Display() error {
	d.setMemoryArea(0, 0, d.width-1, d.height-1)
	for j := int16(0); j < d.height; j++ {
		d.setMemoryPointer(0, j)
		d.SendCommand(WRITE_RAM)
		for i := int16(0); i < d.width/8; i++ {
			d.SendData(d.buffer[i+j*(d.width/8)])
		}
	}

	d.SendCommand(DISPLAY_UPDATE_CONTROL_2)
	d.SendData(0xC4)
	d.SendCommand(MASTER_ACTIVATION)
	d.SendCommand(TERMINATE_FRAME_READ_WRITE)
	return nil
}

// DisplayRect sends only an area of the buffer to the screen.
func (d *Device) DisplayRect(x int16, y int16, width int16, height int16) error {
	if x < 0 || y < 0 || x >= d.width || y >= d.height || width < 0 || height < 0 {
		return errors.New("wrong rectangle")
	}
	x &= 0xF8
	width &= 0xF8
	width = x + width // reuse variables
	if width >= d.width {
		width = d.width
	}
	height = y + height
	if height > d.height {
		height = d.height
	}
	d.setMemoryArea(x, y, width, height)
	x = x / 8
	width = width / 8
	for ; y < height; y++ {
		d.setMemoryPointer(8*x, y)
		d.SendCommand(WRITE_RAM)
		for i := int16(x); i < width; i++ {
			d.SendData(d.buffer[i+y*d.width/8])
		}
	}

	d.SendCommand(DISPLAY_UPDATE_CONTROL_2)
	d.SendData(0xC4)
	d.SendCommand(MASTER_ACTIVATION)
	d.SendCommand(TERMINATE_FRAME_READ_WRITE)
	return nil
}

// ClearDisplay erases the device SRAM
func (d *Device) ClearDisplay() {
	d.setMemoryArea(0, 0, d.width-1, d.height-1)
	d.setMemoryPointer(0, 0)
	d.SendCommand(WRITE_RAM)
	for i := uint32(0); i < d.bufferLength; i++ {
		d.SendData(0xFF)
	}
	d.Display()
}

// setMemoryArea sets the area of the display that will be updated
func (d *Device) setMemoryArea(x0 int16, y0 int16, x1 int16, y1 int16) {
	d.SendCommand(SET_RAM_X_ADDRESS_START_END_POSITION)
	d.SendData(uint8((x0 >> 3) & 0xFF))
	d.SendData(uint8((x1 >> 3) & 0xFF))
	d.SendCommand(SET_RAM_Y_ADDRESS_START_END_POSITION)
	d.SendData(uint8(y0 & 0xFF))
	d.SendData(uint8((y0 >> 8) & 0xFF))
	d.SendData(uint8(y1 & 0xFF))
	d.SendData(uint8((y1 >> 8) & 0xFF))
}

// setMemoryPointer moves the internal pointer to the speficied coordinates
func (d *Device) setMemoryPointer(x int16, y int16) {
	d.SendCommand(SET_RAM_X_ADDRESS_COUNTER)
	d.SendData(uint8((x >> 3) & 0xFF))
	d.SendCommand(SET_RAM_Y_ADDRESS_COUNTER)
	d.SendData(uint8(y & 0xFF))
	d.SendData(uint8((y >> 8) & 0xFF))
	d.WaitUntilIdle()
}

// WaitUntilIdle waits until the display is ready
func (d *Device) WaitUntilIdle() {
	for d.busy.Get() {
		time.Sleep(100 * time.Millisecond)
	}
}

// IsBusy returns the busy status of the display
func (d *Device) IsBusy() bool {
	return d.busy.Get()
}

// ClearBuffer sets the buffer to 0xFF (white)
func (d *Device) ClearBuffer() {
	for i := uint32(0); i < d.bufferLength; i++ {
		d.buffer[i] = 0xFF
	}
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return d.width, d.height
}
