package sx127x

import (
	"machine"
)

// RadioControl for boards that are connected using normal pins.
type RadioControl struct {
	nssPin, dio0Pin, dio1Pin machine.Pin
}

func NewRadioControl(nssPin, dio0Pin, dio1Pin machine.Pin) *RadioControl {
	return &RadioControl{
		nssPin:  nssPin,
		dio0Pin: dio0Pin,
		dio1Pin: dio1Pin,
	}
}

// SetNss sets the NSS line aka chip select for SPI.
func (rc *RadioControl) SetNss(state bool) error {
	rc.nssPin.Set(state)
	return nil
}

// Init() configures whatever needed for sx127x radio control
func (rc *RadioControl) Init() error {
	rc.nssPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rc.dio0Pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	rc.dio1Pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	return nil
}

// add interrupt handlers for Radio IRQs for pins
func (rc *RadioControl) SetupInterrupts(handler func()) error {
	irqHandler = handler

	// Setup DIO0 interrupt Handling
	if err := rc.dio0Pin.SetInterrupt(machine.PinRising, handleInterrupt); err != nil {
		return err
	}

	// Setup DIO1 interrupt Handling
	if err := rc.dio1Pin.SetInterrupt(machine.PinRising, handleInterrupt); err != nil {
		return err
	}

	return nil
}

var irqHandler func()

func handleInterrupt(machine.Pin) {
	irqHandler()
}
