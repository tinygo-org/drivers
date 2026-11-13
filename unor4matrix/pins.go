package unor4matrix

import pin "tinygo.org/x/drivers/internal/pin"

const numPins = 11

// charlieplexPin can drive a line or put it in high-impedance input mode.
// Set and Float must not block or allocate. Float must also initialize the pin.
type charlieplexPin struct {
	Set   pin.OutputFunc
	Float func()
}
