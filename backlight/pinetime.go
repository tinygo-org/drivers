package backlight

import "machine"

type PineTimeDriver struct {
	blLowPin  machine.Pin
	blMidPin  machine.Pin
	blHighPin machine.Pin
}

func NewPineTimeDriver(lowPin, midPin, highPin machine.Pin) PineTimeDriver {
	lowPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	midPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	highPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return PineTimeDriver{blLowPin: lowPin, blMidPin: midPin, blHighPin: highPin}
}

// Not sure if you can combine the three pins to get intermediate brightness levels.s
func (b PineTimeDriver) SetBrightness(brightness uint8) {
	if brightness < 85 {
		b.blLowPin.Low()
		b.blMidPin.High()
		b.blHighPin.High()
	} else if brightness < 170 {
		b.blLowPin.High()
		b.blMidPin.Low()
		b.blHighPin.High()
	} else {
		b.blLowPin.High()
		b.blMidPin.High()
		b.blHighPin.Low()
	}
}
