// Package microbitmatrix implements a driver for the BBC micro:bit's LED matrix.
//
// Schematic: https://github.com/bbcmicrobit/hardware/blob/master/SCH_BBC-Microbit_V1.3B.pdf
//
package microbitmatrix // import "tinygo.org/x/drivers/microbitmatrix"

import (
	"image/color"
)

type Config struct {
	Rotation uint8
}

// New returns a new microbitmatrix driver.
func New() Device {
	return Device{}
}

// SetRotation changes the rotation of the LED matrix
func (d *Device) SetRotation(rotation uint8) {
	d.rotation = rotation % 4
}

// SetPixel modifies the internal buffer in a single pixel.
func (d *Device) SetPixel(x int16, y int16, c color.RGBA) {
	if x < 0 || x >= 5 || y < 0 || y >= 5 {
		return
	}
	if c.R != 0 || c.G != 0 || c.B != 0 {
		d.buffer[matrixRotations[d.rotation][x][y][0]][matrixRotations[d.rotation][x][y][1]] = true
	} else {
		d.buffer[matrixRotations[d.rotation][x][y][0]][matrixRotations[d.rotation][x][y][1]] = false
	}
}

// GetPixel returns if the specific pixels is enabled
func (d *Device) GetPixel(x int16, y int16) bool {
	if x < 0 || x >= 5 || y < 0 || y >= 5 {
		return false
	}
	return d.buffer[matrixRotations[d.rotation][x][y][0]][matrixRotations[d.rotation][x][y][1]]
}
