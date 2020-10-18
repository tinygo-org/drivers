// Package pcd8544 implements a driver for the PCD8544 48x84 pixels matrix LCD, used in Nokia's 5110 and 3310 phones.
//
// Datasheet: http://eia.udg.edu/~forest/PCD8544_1.pdf
//
package pcd8544 // import "tinygo.org/x/drivers/pcd8544"

import (
	"errors"
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

// Device wraps an SPI connection.
type Device struct {
	bus        drivers.SPI
	dcPin      machine.Pin
	rstPin     machine.Pin
	scePin     machine.Pin
	buffer     []byte
	width      int16
	height     int16
	bufferSize int16
}

type Config struct {
	Width  int16
	Height int16
}

// New creates a new PCD8544 connection. The SPI bus must already be configured.
func New(bus drivers.SPI, dcPin, rstPin, scePin machine.Pin) *Device {
	return &Device{
		bus:    bus,
		dcPin:  dcPin,
		rstPin: rstPin,
		scePin: scePin,
	}
}

// Configure initializes the display with default configuration
func (d *Device) Configure(cfg Config) {
	if cfg.Width != 0 {
		d.width = cfg.Width
	} else {
		d.width = 84
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	} else {
		d.height = 48
	}
	d.bufferSize = d.width * d.height / 8
	d.buffer = make([]byte, d.bufferSize)

	d.rstPin.Low()
	time.Sleep(100 * time.Nanosecond)
	d.rstPin.High()
	d.SendCommand(FUNCTIONSET | EXTENDEDINSTRUCTION) // H = 1
	d.SendCommand(SETVOP | 0x3f)                     // 0x3f : Vop6 = 0, Vop5 to Vop0 = 1
	d.SendCommand(SETTEMP | 0x03)                    // Experimentally determined
	d.SendCommand(SETBIAS | 0x03)                    // Experimentally determined
	d.SendCommand(FUNCTIONSET)                       // H = 0
	d.SendCommand(DISPLAYCONTROL | DISPLAYNORMAL)
}

// ClearBuffer clears the image buffer
func (d *Device) ClearBuffer() {
	d.buffer = make([]byte, d.bufferSize)
}

// ClearDisplay clears the image buffer and clear the display
func (d *Device) ClearDisplay() {
	d.ClearBuffer()
	d.Display()
}

// Display sends the whole buffer to the screen
func (d *Device) Display() error {
	d.SendCommand(FUNCTIONSET) // H = 0
	d.SendCommand(SETXADDR)
	d.SendCommand(SETYADDR)

	for i := int16(0); i < d.bufferSize; i++ {
		d.SendData(d.buffer[i])
	}
	return nil
}

// sendDataCommand sends image data or a command to the screen
func (d *Device) sendDataCommand(isCommand bool, data uint8) {
	if isCommand {
		d.dcPin.Low()
	} else {
		d.dcPin.High()
	}
	d.scePin.Low()
	d.bus.Transfer(data)
	d.scePin.High()
}

// SetPixel enables or disables a pixel in the buffer
// color.RGBA{0, 0, 0, 255} is consider transparent, anything else
// with enable a pixel on the screen
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return
	}
	byteIndex := x + (y/8)*d.width
	if c.R != 0 || c.G != 0 || c.B != 0 {
		d.buffer[byteIndex] |= 1 << uint8(y%8)
	} else {
		d.buffer[byteIndex] &^= 1 << uint8(y%8)
	}
}

// GetPixel returns if the specified pixel is on (true) or off (false)
func (d *Device) GetPixel(x int16, y int16) bool {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return false
	}
	byteIndex := x + (y/8)*d.width
	return (d.buffer[byteIndex] >> uint8(y%8) & 0x1) == 1
}

// SetBuffer changes the whole buffer at once
func (d *Device) SetBuffer(buffer []byte) error {
	if int16(len(buffer)) != d.bufferSize {
		//return ErrBuffer
		return errors.New("wrong size buffer")
	}
	for i := int16(0); i < d.bufferSize; i++ {
		d.buffer[i] = buffer[i]
	}
	return nil
}

// SendCommand sends a command to the display
func (d *Device) SendCommand(command uint8) {
	d.sendDataCommand(true, command)
}

// SendData sends a data byte to the display
func (d *Device) SendData(data uint8) {
	d.sendDataCommand(false, data)
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return d.width, d.height
}
