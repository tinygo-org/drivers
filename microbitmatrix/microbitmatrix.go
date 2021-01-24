// +build microbit

// Package microbitmatrix implements a driver for the BBC micro:bit's LED matrix.
//
// Schematic: https://github.com/bbcmicrobit/hardware/blob/master/SCH_BBC-Microbit_V1.3B.pdf
//
package microbitmatrix // import "tinygo.org/x/drivers/microbitmatrix"

import (
	"image/color"
	"machine"
	"time"
)

var matrixRotations = [4][5][5][2]uint8{
	{ // 0
		{{0, 0}, {1, 3}, {0, 1}, {1, 4}, {0, 2}},
		{{2, 3}, {2, 4}, {2, 5}, {2, 6}, {2, 7}},
		{{1, 1}, {0, 8}, {1, 2}, {2, 8}, {1, 0}},
		{{0, 7}, {0, 6}, {0, 5}, {0, 4}, {0, 3}},
		{{2, 2}, {1, 6}, {2, 0}, {1, 5}, {2, 1}},
	},
	{ // 90 CCW
		{{0, 2}, {2, 7}, {1, 0}, {0, 3}, {2, 1}},
		{{1, 4}, {2, 6}, {2, 8}, {0, 4}, {1, 5}},
		{{0, 1}, {2, 5}, {1, 2}, {0, 5}, {2, 0}},
		{{1, 3}, {2, 4}, {0, 8}, {0, 6}, {1, 6}},
		{{0, 0}, {2, 3}, {1, 1}, {0, 7}, {2, 2}},
	},
	{ // 180
		{{2, 1}, {1, 5}, {2, 0}, {1, 6}, {2, 2}},
		{{0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7}},
		{{1, 0}, {2, 8}, {1, 2}, {0, 8}, {1, 1}},
		{{2, 7}, {2, 6}, {2, 5}, {2, 4}, {2, 3}},
		{{0, 2}, {1, 4}, {0, 1}, {1, 3}, {0, 0}},
	},
	{ // 270
		{{2, 2}, {0, 7}, {1, 1}, {2, 3}, {0, 0}},
		{{1, 6}, {0, 6}, {0, 8}, {2, 4}, {1, 3}},
		{{2, 0}, {0, 5}, {1, 2}, {2, 5}, {0, 1}},
		{{1, 5}, {0, 4}, {2, 8}, {2, 6}, {1, 4}},
		{{2, 1}, {0, 3}, {1, 0}, {2, 7}, {0, 2}},
	},
}

type Config struct {
	Rotation uint8
}

type Device struct {
	pin      [12]machine.Pin
	buffer   [3][9]bool
	rotation uint8
}

// New returns a new microbitmatrix driver.
func New() Device {
	return Device{}
}

// Configure sets up the device.
func (d *Device) Configure(cfg Config) {
	d.SetRotation(cfg.Rotation)

	for i := machine.LED_COL_1; i <= machine.LED_ROW_3; i++ {
		d.pin[i-machine.LED_COL_1] = i
		d.pin[i-machine.LED_COL_1].Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
	d.ClearDisplay()
	d.DisableAll()
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

// Display sends the buffer (if any) to the screen.
func (d *Device) Display() error {
	for row := 0; row < 3; row++ {
		d.DisableAll()
		d.pin[9+row].High()

		for col := 0; col < 9; col++ {
			if d.buffer[row][col] {
				d.pin[col].Low()
			}

		}
		time.Sleep(time.Millisecond * 2)
	}
	return nil
}

// ClearDisplay erases the internal buffer
func (d *Device) ClearDisplay() {
	for row := 0; row < 3; row++ {
		for col := 0; col < 9; col++ {
			d.buffer[row][col] = false
		}
	}
}

// DisableAll disables all the LEDs without modifying the buffer
func (d *Device) DisableAll() {
	for i := machine.LED_COL_1; i <= machine.LED_COL_9; i++ {
		d.pin[i-machine.LED_COL_1].High()
	}
	for i := machine.LED_ROW_1; i <= machine.LED_ROW_3; i++ {
		d.pin[i-machine.LED_COL_1].Low()
	}
}

// EnableAll enables all the LEDs without modifying the buffer
func (d *Device) EnableAll() {
	for i := machine.LED_COL_1; i <= machine.LED_COL_9; i++ {
		d.pin[i-machine.LED_COL_1].Low()
	}
	for i := machine.LED_ROW_1; i <= machine.LED_ROW_3; i++ {
		d.pin[i-machine.LED_COL_1].High()
	}
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return 5, 5
}
