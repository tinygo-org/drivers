//go:build microbit_v2

// Package microbitmatrix implements a driver for the BBC micro:bit version 2 LED matrix.
//
// Schematic: https://github.com/microbit-foundation/microbit-v2-hardware/blob/main/V2.00/MicroBit_V2.0.0_S_schematic.PDF
package microbitmatrix // import "tinygo.org/x/drivers/microbitmatrix"

import (
	"machine"
)

// 4 rotation orientations (0, 90, 180, 270), CW (clock wise)
// 5 rows
// 5 cols
// target coordinates in machine rows (y) and cols (x)
var matrixRotations = [4][5][5][2]uint8{
	{ // 0
		{{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}},
		{{1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}},
		{{2, 0}, {2, 1}, {2, 2}, {2, 3}, {2, 4}},
		{{3, 0}, {3, 1}, {3, 2}, {3, 3}, {3, 4}},
		{{4, 0}, {4, 1}, {4, 2}, {4, 3}, {4, 4}},
	},
	{ // 90 CW
		{{0, 4}, {1, 4}, {2, 4}, {3, 4}, {4, 4}},
		{{0, 3}, {1, 3}, {2, 3}, {3, 3}, {4, 3}},
		{{0, 2}, {1, 2}, {2, 2}, {3, 2}, {4, 2}},
		{{0, 1}, {1, 1}, {2, 1}, {3, 1}, {4, 1}},
		{{0, 0}, {1, 0}, {2, 0}, {3, 0}, {4, 0}},
	},
	{ // 180
		{{4, 4}, {4, 3}, {4, 2}, {4, 1}, {4, 0}},
		{{3, 4}, {3, 3}, {3, 2}, {3, 1}, {3, 0}},
		{{2, 4}, {2, 3}, {2, 2}, {2, 1}, {2, 0}},
		{{1, 4}, {1, 3}, {1, 2}, {1, 1}, {1, 0}},
		{{0, 4}, {0, 3}, {0, 2}, {0, 1}, {0, 0}},
	},
	{ // 270
		{{4, 0}, {3, 0}, {2, 0}, {1, 0}, {0, 0}},
		{{4, 1}, {3, 1}, {2, 1}, {1, 1}, {0, 1}},
		{{4, 2}, {3, 2}, {2, 2}, {1, 2}, {0, 2}},
		{{4, 3}, {3, 3}, {2, 3}, {1, 3}, {0, 3}},
		{{4, 4}, {3, 4}, {2, 4}, {1, 4}, {0, 4}},
	},
}

const (
	ledRows = 5
	ledCols = 5
)

type Device struct {
	pin      [ledCols + ledRows]machine.Pin
	buffer   [ledRows][ledCols]int8
	rotation uint8
}

func (d *Device) assignPins() {
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

	for i := 0; i < len(d.pin); i++ {
		d.pin[i].Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
}
