package unor4matrix

import pin "tinygo.org/x/drivers/internal/pin"

const numPins = 11

// CharlieplexPin can drive a line or put it in high-impedance input mode.
// Set and Float must not block or allocate. Float must also initialize the pin.
type CharlieplexPin struct {
	Set   pin.OutputFunc
	Float func()
}

// New returns a device for P0_03, P0_04, P0_11, P0_12, P0_13, P0_15,
// P2_04, P2_05, P2_06, P2_12 and P2_13, in that order.
func New(pins [numPins]CharlieplexPin) Device {
	return Device{pins: pins}
}
