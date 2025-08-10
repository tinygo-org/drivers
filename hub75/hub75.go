// Package hub75 implements a driver for the HUB75 LED matrix.
//
// Guide: https://cdn-learn.adafruit.com/downloads/pdf/32x16-32x32-rgb-led-matrix.pdf
// This driver was inspired by https://github.com/2dom/PxMatrix
package hub75 // import "tinygo.org/x/drivers/hub75"

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
)

type Config struct {
	Width      int16
	Height     int16
	ColorDepth uint16
	RowPattern int16
	Brightness uint8
	FastUpdate bool
}

type Device struct {
	bus               drivers.SPI
	a                 machine.Pin
	b                 machine.Pin
	c                 machine.Pin
	d                 machine.Pin
	oe                machine.Pin
	lat               machine.Pin
	width             int16
	height            int16
	brightness        uint8
	fastUpdate        bool
	colorDepth        uint16
	colorStep         uint16
	colorHalfStep     uint16
	colorThirdStep    uint16
	colorTwoThirdStep uint16
	rowPattern        int16
	rowsPerBuffer     int16
	panelWidth        int16
	panelWidthBytes   int16
	pixelCounter      uint32
	lineCounter       uint32
	patternColorBytes uint8
	rowSetsPerBuffer  uint8
	sendBufferSize    uint16
	rowOffset         []uint32
	buffer            [][]uint8 // [ColorDepth][(width * height * 3(rgb)) / 8]uint8
	displayColor      uint16
}

// New returns a new HUB75 driver. Pass in a fully configured SPI bus.
func New(b drivers.SPI, latPin, oePin, aPin, bPin, cPin, dPin machine.Pin) Device {
	aPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	bPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	cPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	dPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	oePin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	latPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	return Device{
		bus: b,
		a:   aPin,
		b:   bPin,
		c:   cPin,
		d:   dPin,
		oe:  oePin,
		lat: latPin,
	}
}

// Configure sets up the device.
func (d *Device) Configure(cfg Config) {
	if cfg.Width != 0 {
		d.width = cfg.Width
	} else {
		d.width = 64
	}
	if cfg.Height != 0 {
		d.height = cfg.Height
	} else {
		d.height = 32
	}
	if cfg.ColorDepth != 0 {
		d.colorDepth = cfg.ColorDepth
	} else {
		d.colorDepth = 8
	}
	if cfg.RowPattern != 0 {
		d.rowPattern = cfg.RowPattern
	} else {
		d.rowPattern = 16
	}
	if cfg.Brightness != 0 {
		d.brightness = cfg.Brightness
	} else {
		d.brightness = 255
	}

	d.fastUpdate = cfg.FastUpdate
	d.rowsPerBuffer = d.height / 2
	d.panelWidth = 1
	d.panelWidthBytes = (d.width / d.panelWidth) / 8
	d.rowOffset = make([]uint32, d.height)
	d.patternColorBytes = uint8((d.height / d.rowPattern) * (d.width / 8))
	d.rowSetsPerBuffer = uint8(d.rowsPerBuffer / d.rowPattern)
	d.sendBufferSize = uint16(d.patternColorBytes) * 3
	d.colorStep = 256 / d.colorDepth
	d.colorHalfStep = d.colorStep / 2
	d.colorThirdStep = d.colorStep / 3
	d.colorTwoThirdStep = 2 * d.colorThirdStep
	d.buffer = make([][]uint8, d.colorDepth)
	for i := range d.buffer {
		d.buffer[i] = make([]uint8, (d.width*d.height*3)/8)
	}

	d.colorHalfStep = d.colorStep / 2
	d.colorThirdStep = d.colorStep / 3
	d.colorTwoThirdStep = 2 * d.colorThirdStep

	d.a.Low()
	d.b.Low()
	d.c.Low()
	d.d.Low()
	d.oe.High()

	var i uint32
	for i = 0; i < uint32(d.height); i++ {
		d.rowOffset[i] = (i%uint32(d.rowPattern))*uint32(d.sendBufferSize) + uint32(d.sendBufferSize) - 1
	}
}

// SetPixel modifies the internal buffer in a single pixel.
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	d.fillMatrixBuffer(x, y, c.R, c.G, c.B)
}

// fillMatrixBuffer modifies a pixel in the internal buffer given position and RGB values
func (d *Device) fillMatrixBuffer(x int16, y int16, r uint8, g uint8, b uint8) {
	if x < 0 || x >= d.width || y < 0 || y >= d.height {
		return
	}
	x = d.width - 1 - x

	var offsetR uint32
	var offsetG uint32
	var offsetB uint32

	vertIndexInBuffer := uint8((int32(y) % int32(d.rowsPerBuffer)) / int32(d.rowPattern))
	whichBuffer := uint8(y / d.rowsPerBuffer)
	xByte := x / 8
	whichPanel := uint8(xByte / d.panelWidthBytes)
	inRowByteOffset := uint8(xByte % d.panelWidthBytes)

	offsetR = d.rowOffset[y] - uint32(inRowByteOffset) - uint32(d.panelWidthBytes)*
		(uint32(d.rowSetsPerBuffer)*(uint32(d.panelWidth)*uint32(whichBuffer)+uint32(whichPanel))+uint32(vertIndexInBuffer))
	offsetG = offsetR - uint32(d.patternColorBytes)
	offsetB = offsetG - uint32(d.patternColorBytes)

	bitSelect := uint8(x % 8)

	for c := uint16(0); c < d.colorDepth; c++ {
		colorTresh := uint8(c*d.colorStep + d.colorHalfStep)
		if r > colorTresh {
			d.buffer[c][offsetR] |= 1 << bitSelect
		} else {
			d.buffer[c][offsetR] &^= 1 << bitSelect
		}
		if g > colorTresh {
			d.buffer[(c+d.colorThirdStep)%d.colorDepth][offsetG] |= 1 << bitSelect
		} else {
			d.buffer[(c+d.colorThirdStep)%d.colorDepth][offsetG] &^= 1 << bitSelect
		}
		if b > colorTresh {
			d.buffer[(c+d.colorTwoThirdStep)%d.colorDepth][offsetB] |= 1 << bitSelect
		} else {
			d.buffer[(c+d.colorTwoThirdStep)%d.colorDepth][offsetB] &^= 1 << bitSelect
		}
	}
}

// Display sends the buffer (if any) to the screen.
func (d *Device) Display() error {
	rp := uint16(d.rowPattern)
	for i := uint16(0); i < rp; i++ {
		// FAST UPDATES (only if brightness = 255)
		if d.fastUpdate && d.brightness == 255 {
			d.setMux((i + rp - 1) % rp)
			d.lat.High()
			d.oe.Low()
			d.lat.Low()
			time.Sleep(1 * time.Microsecond)
			d.bus.Tx(d.buffer[d.displayColor][i*d.sendBufferSize:(i+1)*d.sendBufferSize], nil)
			time.Sleep(10 * time.Microsecond)
			d.oe.High()

		} else { // NO FAST UPDATES
			d.setMux(i)
			d.bus.Tx(d.buffer[d.displayColor][i*d.sendBufferSize:(i+1)*d.sendBufferSize], nil)
			d.latch((255 * uint16(d.brightness)) / 255)
		}
	}
	d.displayColor++
	if d.displayColor >= d.colorDepth {
		d.displayColor = 0
	}
	return nil
}

func (d *Device) latch(showTime uint16) {
	d.lat.High()
	d.lat.Low()
	d.oe.Low()
	time.Sleep(time.Duration(showTime) * time.Microsecond)
	d.oe.High()
}

func (d *Device) setMux(value uint16) {
	if (value & 0x01) == 0x01 {
		d.a.High()
	} else {
		d.a.Low()
	}
	if (value & 0x02) == 0x02 {
		d.b.High()
	} else {
		d.b.Low()
	}
	if (value & 0x04) == 0x04 {
		d.c.High()
	} else {
		d.c.Low()
	}
	if (value & 0x08) == 0x08 {
		d.d.High()
	} else {
		d.d.Low()
	}
}

// FlushDisplay flushes the display
func (d *Device) FlushDisplay() {
	var i uint16
	for i = 0; i < d.sendBufferSize; i++ {
		d.bus.Tx([]byte{0x00}, nil)
	}
}

// SetBrightness changes the brightness of the display
func (d *Device) SetBrightness(brightness uint8) {
	d.brightness = brightness
}

// ClearDisplay erases the internal buffer
func (d *Device) ClearDisplay() {
	bufferSize := (d.width * d.height * 3) / 8
	for c := uint16(0); c < d.colorDepth; c++ {
		for j := int16(0); j < bufferSize; j++ {
			d.buffer[c][j] = 0
		}
	}
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return d.width, d.height
}
