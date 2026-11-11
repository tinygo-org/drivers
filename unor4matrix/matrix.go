// Package unor4matrix drives the UNO R4 WiFi's built-in 12x8 LED matrix.
// Display scans a complete frame and blanks the lines before returning.
package unor4matrix

import (
	"errors"
	"image/color"
	"time"
)

// Device stores one drawing buffer. Use it from one application goroutine.
type Device struct {
	pins       [numPins]CharlieplexPin
	frame      framebuffer
	rotation   uint8
	configured bool
}

// Rotation values are clockwise quarter turns.
const (
	RotationNormal = 0
	Rotation90     = 1
	Rotation180    = 2
	Rotation270    = 3
)

// refreshPeriod matches the Arduino matrix driver's 10 kHz scan timer.
// https://github.com/arduino/ArduinoCore-renesas/blob/424e86eff92d37f72123c2b641dd8bbf06a38b47/libraries/Arduino_LED_Matrix/src/Arduino_LED_Matrix.h#L172
const refreshPeriod = 100 * time.Microsecond

// ErrNotConfigured reports Display called before successful Configure.
var ErrNotConfigured = errors.New("unor4matrix: not configured")

// ErrMissingPinFunc reports a missing Set or Float function.
var ErrMissingPinFunc = errors.New("unor4matrix: all eleven lines require Set and Float")

// Configure floats all lines and retains the stored image.
// The caller exclusively owns the lines and serializes display operations.
func (d *Device) Configure() error {
	for _, pin := range d.pins {
		if pin.Set == nil || pin.Float == nil {
			return ErrMissingPinFunc
		}
	}
	for _, pin := range d.pins {
		pin.Float()
	}
	d.configured = true
	return nil
}

// SetPixel edits one logical pixel. Nonzero RGB is on. Alpha is ignored.
// Out-of-bounds writes are ignored. Call Display to show changes.
func (d *Device) SetPixel(x, y int16, c color.RGBA) {
	index, ok := d.pixelIndex(x, y)
	if !ok {
		return
	}
	d.frame.set(index, c.R != 0 || c.G != 0 || c.B != 0)
}

// GetPixel reports whether a buffered pixel is on. Out-of-bounds reads return false.
func (d *Device) GetPixel(x, y int16) bool {
	index, ok := d.pixelIndex(x, y)
	if !ok {
		return false
	}
	return d.frame.get(index)
}

// SetRotation sets clockwise quarter turns modulo four and retains the buffer.
func (d *Device) SetRotation(rotation uint8) { d.rotation = rotation % 4 }

// Size returns the drawing area for the current rotation.
func (d *Device) Size() (width, height int16) {
	if d.rotation == Rotation90 || d.rotation == Rotation270 {
		return Height, Width
	}
	return Width, Height
}

func (d *Device) pixelIndex(x, y int16) (int, bool) {
	w, h := d.Size()
	if x < 0 || x >= w || y < 0 || y >= h {
		return 0, false
	}
	switch d.rotation {
	case Rotation90:
		x, y = Width-1-y, x
	case Rotation180:
		x, y = Width-1-x, Height-1-y
	case Rotation270:
		x, y = y, Height-1-x
	}
	return int(y)*Width + int(x), true
}

// Display scans one frame and blanks the lines. Call repeatedly to refresh.
// On RA4M1, use tasks. With -scheduler=none, refresh may fall to about 10 Hz.
func (d *Device) Display() error {
	if !d.configured {
		return ErrNotConfigured
	}
	for index, pair := range ledMapping {
		on := d.frame.get(index)
		if on {
			d.pins[pair.cathode].Set.Low()
			d.pins[pair.anode].Set.High()
		}
		// Yield between slots so other goroutines can run.
		// They may extend the active LED's on-time.
		time.Sleep(refreshPeriod)
		if on {
			d.pins[pair.anode].Float()
			d.pins[pair.cathode].Float()
		}
	}
	return nil
}

// ClearDisplay clears the drawing buffer. Call Display to show the change.
func (d *Device) ClearDisplay() { d.frame.clear() }

// LoadFrame edits the drawing buffer using Arduino row-major, MSB-first words.
// Rotation does not affect the word layout. Call Display to show the image.
func (d *Device) LoadFrame(words [3]uint32) {
	d.frame.clear()
	for index := 0; index < Width*Height; index++ {
		if words[index/32]&(uint32(1)<<(31-index%32)) != 0 {
			d.frame.set(index, true)
		}
	}
}
