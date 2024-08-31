// Package epd1in54 implements a driver for Waveshare 1.54in black and white e-paper device.
//
// Derived from:
//
//	https://github.com/tinygo-org/drivers/tree/master/waveshare-epd
//	https://github.com/waveshare/e-Paper/blob/master/Arduino/epd1in54_V2/epd1in54_V2.cpp
//
// Datasheet: https://www.waveshare.com/w/upload/e/e5/1.54inch_e-paper_V2_Datasheet.pdf
package epd1in54

import (
	"image/color"
	"machine"
	"time"
)

type Config struct {
	Width        int16
	Height       int16
	LogicalWidth int16
	Rotation     Rotation
}

type Device struct {
	bus  machine.SPI
	cs   machine.Pin
	dc   machine.Pin
	rst  machine.Pin
	busy machine.Pin

	buffer   []uint8
	rotation Rotation
}

type Rotation uint8

var fullRefresh = [159]uint8{
	0x80, 0x48, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x40, 0x48, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x80, 0x48, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x40, 0x48, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0xA, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x8, 0x1, 0x0, 0x8, 0x1, 0x0, 0x2,
	0xA, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x0, 0x0, 0x0,
	0x22, 0x17, 0x41, 0x0, 0x32, 0x20,
}

var partialRefresh = [159]uint8{
	0x0, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x80, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x40, 0x40, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x80, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0xF, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x1, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x0, 0x0, 0x0,
	0x02, 0x17, 0x41, 0xB0, 0x32, 0x28,
}

// New returns a new epd1in54 driver. Pass in a fully configured SPI bus.
func New(bus machine.SPI, csPin, dcPin, rstPin, busyPin machine.Pin) Device {
	return Device{
		buffer: make([]uint8, (uint32(Width)*uint32(Height))/8),
		bus:    bus,
		cs:     csPin,
		dc:     dcPin,
		rst:    rstPin,
		busy:   busyPin,
	}
}

func (d *Device) LDirInit(cfg Config) {
	d.cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.rst.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.dc.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.busy.Configure(machine.PinConfig{Mode: machine.PinInput})

	d.bus.Configure(machine.SPIConfig{
		Frequency: 2000000,
		Mode:      0,
		LSBFirst:  false,
	})

	d.Reset()
	d.WaitUntilIdle()

	d.SendCommand(0x12)
	d.WaitUntilIdle()

	d.SendCommand(0x01)
	d.SendData(0xC7)
	d.SendData(0x00)
	d.SendData(0x00)

	d.SendCommand(0x11)
	d.SendData(0x03)

	d.SendCommand(0x44)
	/* x point must be the multiple of 8 or the last 3 bits will be ignored */
	d.SendData((0 >> 3) & 0xFF)
	d.SendData((199 >> 3) & 0xFF)

	d.SendCommand(0x45)
	d.SendData(0 & 0xFF)
	d.SendData((0 >> 8) & 0xFF)
	d.SendData(199 & 0xFF)
	d.SendData((199 >> 8) & 0xFF)

	d.SendCommand(0x3C)
	d.SendData(0x01)

	d.SendCommand(0x18)
	d.SendData(0x80)

	d.SendCommand(0x22)
	d.SendData(0xB1)

	d.SendCommand(0x20)

	d.SendCommand(0x4E)
	d.SendData(0x00)

	d.SendCommand(0x4F)
	d.SendData(0xC7)
	d.SendData(0x00)

	d.WaitUntilIdle()
	d.setLUT(fullRefresh)
}

func (d *Device) HDirInit(cfg Config) {
	d.cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.rst.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.dc.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d.busy.Configure(machine.PinConfig{Mode: machine.PinInput})

	d.bus.Configure(machine.SPIConfig{
		Frequency: 2000000,
		Mode:      0,
		LSBFirst:  false,
	})

	d.Reset()
	d.WaitUntilIdle()

	d.SendCommand(0x12)
	d.WaitUntilIdle()

	d.SendCommand(0x01)
	d.SendData(0xC7)
	d.SendData(0x00)
	d.SendData(0x01)

	d.SendCommand(0x11)
	d.SendData(0x01)

	d.SendCommand(0x44)
	d.SendData(0x00)
	d.SendData(0x18)

	d.SendCommand(0x45)
	d.SendData(0xC7)
	d.SendData(0x00)
	d.SendData(0x00)
	d.SendData(0x00)

	d.SendCommand(0x3C)
	d.SendData(0x01)

	d.SendCommand(0x18)
	d.SendData(0x80)

	d.SendCommand(0x22)
	d.SendData(0xB1)

	d.SendCommand(0x20)

	d.SendCommand(0x4E)
	d.SendData(0x00)

	d.SendCommand(0x4F)
	d.SendData(0xC7)
	d.SendData(0x00)

	d.WaitUntilIdle()
	d.setLUT(fullRefresh)
}

func (d *Device) setLUT(lut [159]uint8) {
	d.SendCommand(0x32)
	for i := 0; i < 153; i++ {
		d.SendData(lut[i])
	}
	d.WaitUntilIdle()

	d.SendCommand(0x3F)
	d.SendData(lut[153])

	d.SendCommand(0x03)
	d.SendData(lut[154])

	d.SendCommand(0x04)
	d.SendData(lut[155])
	d.SendData(lut[156])
	d.SendData(lut[157])

	d.SendCommand(0x2C)
	d.SendData(lut[158])
}

// Reset resets the display.
func (d *Device) Reset() {
	d.rst.High()
	time.Sleep(20 * time.Millisecond)
	d.rst.Low()
	time.Sleep(5 * time.Millisecond)
	d.rst.High()
	time.Sleep(20 * time.Millisecond)
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
// The display have 2 colors: black and white
// We use RGBA(0,0,0, 255) as white (transparent)
// Anything else as black
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	x, y = d.xy(x, y)
	if x < 0 || x >= Width || y < 0 || y >= Height {
		return
	}
	byteIndex := (uint32(x) + uint32(y)*uint32(Width)) / 8
	if c.R == 0 && c.G == 0 && c.B == 0 { // TRANSPARENT / WHITE
		d.buffer[byteIndex] |= 0x80 >> uint8(x%8)
	} else { // WHITE / EMPTY
		d.buffer[byteIndex] &^= 0x80 >> uint8(x%8)
	}
}

func (d *Device) DisplayImage(image []uint8) {
	var w, h int
	if Width%8 == 0 {
		w = int(Width / 8)
	} else {
		w = int(Width/8 + 1)
	}
	h = int(Height)

	d.SendCommand(0x24)
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			d.SendData(image[i+j*w])
		}
	}

	d.SendCommand(0x26)
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			d.SendData(image[i+j*w])
		}
	}

	d.displayFrame()
}

func (d *Device) Display() error {
	var w, h int
	if Width%8 == 0 {
		w = int(Width / 8)
	} else {
		w = int(Width/8 + 1)
	}
	h = int(Height)

	d.SendCommand(0x24)
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			x := i + j*w
			d.SendData(d.buffer[x])
		}
	}

	d.SendCommand(0x26)
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			x := i + j*w
			d.SendData(d.buffer[x])
		}
	}

	d.displayFrame()

	return nil
}

func (d *Device) displayFrame() {
	d.SendCommand(0x22)
	d.SendData(0xC7)
	d.SendCommand(0x20)
	d.WaitUntilIdle()
}

func (d *Device) Clear() {
	var w, h int
	if Width%8 == 0 {
		w = int(Width / 8)
	} else {
		w = int(Width/8 + 1)
	}
	h = int(Height)

	d.SendCommand(0x24)
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			d.SendData(0xff)
		}
	}

	d.SendCommand(0x26)
	for j := 0; j < h; j++ {
		for i := 0; i < w; i++ {
			d.SendData(0xff)
		}
	}

	d.displayFrame()
}

// WaitUntilIdle waits until the display is ready
func (d *Device) WaitUntilIdle() {
	for d.busy.Get() {
		time.Sleep(100 * time.Millisecond)
	}
	time.Sleep(200 * time.Millisecond)
}

// IsBusy returns the busy status of the display
func (d *Device) IsBusy() bool {
	return d.busy.Get()
}

// ClearBuffer sets the buffer to 0xFF (white)
func (d *Device) ClearBuffer() {
	for i := 0; i < len(d.buffer); i++ {
		d.buffer[i] = 0xFF
	}
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	if d.rotation == ROTATION_90 || d.rotation == ROTATION_270 {
		return Height, Width
	}
	return Width, Height
}

// SetRotation changes the rotation (clock-wise) of the device
func (d *Device) SetRotation(rotation Rotation) {
	d.rotation = rotation
}

// xy chages the coordinates according to the rotation
func (d *Device) xy(x, y int16) (int16, int16) {
	switch d.rotation {
	case NO_ROTATION:
		return x, y
	case ROTATION_90:
		return Width - y - 1, x
	case ROTATION_180:
		return Width - x - 1, Height - y - 1
	case ROTATION_270:
		return y, Height - x - 1
	}
	return x, y
}

func (d *Device) Sleep() {
	d.SendCommand(0x10)
	d.SendData(0x01)
	time.Sleep(200 * time.Millisecond)

	d.rst.Low()
}
