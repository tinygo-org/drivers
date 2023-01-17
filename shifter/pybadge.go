//go:build pybadge

package shifter

import "machine"

const (
	BUTTON_LEFT   = 0
	BUTTON_UP     = 1
	BUTTON_DOWN   = 2
	BUTTON_RIGHT  = 3
	BUTTON_SELECT = 4
	BUTTON_START  = 5
	BUTTON_A      = 6
	BUTTON_B      = 7
)

// NewButtons returns a new shifter device for the buttons on an AdaFruit PyBadge
func NewButtons() Device {
	return Device{
		latch: machine.BUTTON_LATCH,
		clk:   machine.BUTTON_CLK,
		out:   machine.BUTTON_OUT,
		Pins:  make([]ShiftPin, int(EIGHT_BITS)),
		bits:  EIGHT_BITS,
	}
}

// ReadInput returns the latest input readings from the PyBadge.
func (d *Device) ReadInput() (uint8, error) {
	return d.Read8Input()
}
