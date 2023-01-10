//go:build microbit

// Package microbitmatrix implements a driver for the BBC micro:bit's LED matrix.
//
// Schematic: https://github.com/bbcmicrobit/hardware/blob/master/SCH_BBC-Microbit_V1.3B.pdf
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
		{{0, 0}, {1, 3}, {0, 1}, {1, 4}, {0, 2}},
		{{2, 3}, {2, 4}, {2, 5}, {2, 6}, {2, 7}},
		{{1, 1}, {0, 8}, {1, 2}, {2, 8}, {1, 0}},
		{{0, 7}, {0, 6}, {0, 5}, {0, 4}, {0, 3}},
		{{2, 2}, {1, 6}, {2, 0}, {1, 5}, {2, 1}},
	},
	{ // 90 CW
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

const (
	ledRows = 3
	ledCols = 9
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
	d.pin[5] = machine.LED_COL_6
	d.pin[6] = machine.LED_COL_7
	d.pin[7] = machine.LED_COL_8
	d.pin[8] = machine.LED_COL_9

	d.pin[9] = machine.LED_ROW_1
	d.pin[10] = machine.LED_ROW_2
	d.pin[11] = machine.LED_ROW_3

	for i := 0; i < len(d.pin); i++ {
		d.pin[i].Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
}
