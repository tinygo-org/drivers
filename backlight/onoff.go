package backlight

import "machine"

type OnOffDriver struct {
	blPin machine.Pin
}

func NewOnOffDriver(pin machine.Pin) OnOffDriver {
	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return OnOffDriver{blPin: pin}
}

func (b OnOffDriver) SetBrightness(brightness uint8) {
	if brightness < 128 {
		b.blPin.Low()
	} else {
		b.blPin.High()
	}
}
