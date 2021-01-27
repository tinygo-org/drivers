// +build microbit_v2

// Package microbitmatrix implements a driver for the BBC micro:bit version 2 LED matrix.
//
// Schematic:
//
package microbitmatrix // import "tinygo.org/x/drivers/microbitmatrix"

import (
	"machine"
	"time"
)

var matrixRotations = [4][5][5][2]uint8{
	{ // 0
		{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}},
		{{0, 1}, {1, 1}, {2, 1}, {3, 1}, {4, 1}},
		{{0, 2}, {1, 2}, {2, 2}, {3, 2}, {4, 2}},
		{{0, 3}, {1, 3}, {2, 3}, {3, 3}, {4, 3}},
		{{0, 4}, {1, 4}, {2, 4}, {3, 4}, {4, 4}},
	},
	{ // 90 CCW
		{{4, 0}, {4, 1}, {4, 2}, {4, 3}, {4, 4}},
		{{3, 0}, {3, 1}, {3, 2}, {3, 3}, {3, 4}},
		{{2, 0}, {2, 1}, {2, 2}, {2, 3}, {2, 4}},
		{{1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}},
		{{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}},
	},
	{ // 180
		{{4, 4}, {3, 4}, {2, 4}, {1, 4}, {0, 4}},
		{{4, 3}, {3, 3}, {2, 3}, {1, 3}, {0, 3}},
		{{4, 2}, {3, 2}, {2, 2}, {1, 2}, {0, 2}},
		{{4, 1}, {3, 1}, {2, 1}, {1, 1}, {0, 1}},
		{{4, 0}, {3, 0}, {2, 0}, {1, 0}, {0, 0}},
	},
	{ // 270
		{{0, 4}, {0, 3}, {0, 2}, {0, 1}, {0, 0}},
		{{1, 4}, {1, 3}, {1, 2}, {1, 1}, {1, 0}},
		{{2, 4}, {2, 3}, {2, 2}, {2, 1}, {2, 0}},
		{{3, 4}, {3, 3}, {3, 2}, {3, 1}, {3, 0}},
		{{4, 4}, {4, 3}, {4, 2}, {4, 1}, {4, 0}},
	},
}

type Device struct {
	pin      [10]machine.Pin
	buffer   [5][5]bool
	rotation uint8
}

// Configure sets up the device.
func (d *Device) Configure(cfg Config) {
	d.SetRotation(cfg.Rotation)

	d.pin[0] = machine.LED_COL_1
	d.pin[1] = machine.LED_COL_2
	d.pin[2] = machine.LED_COL_3
	d.pin[3] = machine.LED_COL_4
	d.pin[4] = machine.LED_COL_5

	d.pin[5] = machine.LED_ROW_1
	d.pin[6] = machine.LED_ROW_2
	d.pin[7] = machine.LED_ROW_3
	d.pin[8] = machine.LED_ROW_4
	d.pin[9] = machine.LED_ROW_5

	for i := 0; i < 10; i++ {
		d.pin[i].Configure(machine.PinConfig{Mode: machine.PinOutput})
	}

	d.ClearDisplay()
	d.DisableAll()
}

// Display sends the buffer (if any) to the screen.
func (d *Device) Display() error {
	for x := 0; x < 5; x++ {
		d.DisableAll()
		d.pin[x].Low()

		for y := 0; y < 5; y++ {
			if d.buffer[x][y] {
				d.pin[5+y].High()
			} else {
				d.pin[5+y].Low()
			}

		}
		time.Sleep(time.Millisecond * 4)
	}
	return nil
}

// ClearDisplay erases the internal buffer
func (d *Device) ClearDisplay() {
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			d.buffer[row][col] = false
		}
	}
}

// DisableAll disables all the LEDs without modifying the buffer
func (d *Device) DisableAll() {
	for i := 0; i < 5; i++ {
		d.pin[i].High()
		d.pin[5+i].Low()
	}
}

// EnableAll enables all the LEDs without modifying the buffer
func (d *Device) EnableAll() {
	for i := 0; i < 5; i++ {
		d.pin[i].Low()
		d.pin[5+i].High()
	}
}

// Size returns the current size of the display.
func (d *Device) Size() (w, h int16) {
	return 5, 5
}
