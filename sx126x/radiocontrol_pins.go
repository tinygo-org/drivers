//go:build !stm32wlx

package sx126x

import (
	"machine"

	"time"
)

// RadioControl for boards that are connected using normal pins.
type RadioControl struct {
	nssPin, busyPin, dio1Pin   machine.Pin
	rxPin, txLowPin, txHighPin machine.Pin
}

func NewRadioControl(nssPin, busyPin, dio1Pin,
	rxPin, txLowPin, txHighPin machine.Pin) *RadioControl {
	return &RadioControl{
		nssPin:    nssPin,
		busyPin:   busyPin,
		dio1Pin:   dio1Pin,
		rxPin:     rxPin,
		txLowPin:  txLowPin,
		txHighPin: txHighPin,
	}
}

// SetNss sets the NSS line aka chip select for SPI.
func (rc *RadioControl) SetNss(state bool) error {
	rc.nssPin.Set(state)
	return nil
}

// WaitWhileBusy wait until the radio is no longer busy
func (rc *RadioControl) WaitWhileBusy() error {
	count := 100
	for count > 0 {
		if !rc.busyPin.Get() {
			return nil
		}
		count--
		time.Sleep(time.Millisecond)
	}
	return errWaitWhileBusyTimeout
}

// Init() configures whatever needed for sx126x radio control
func (rc *RadioControl) Init() error {
	rc.nssPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rc.busyPin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	return nil
}

// add interrupt handler for Radio IRQs for pins
func (rc *RadioControl) SetupInterrupts(handler func()) error {
	irqHandler = handler

	rc.dio1Pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	if err := rc.dio1Pin.SetInterrupt(machine.PinRising, handleInterrupt); err != nil {
		return errRadioNotFound
	}

	return nil
}

var irqHandler func()

func handleInterrupt(machine.Pin) {
	irqHandler()
}

func (rc *RadioControl) SetRfSwitchMode(mode int) error {
	switch mode {

	case RFSWITCH_RX:
		rc.rxPin.Set(true)
		rc.txLowPin.Set(false)
		rc.txHighPin.Set(false)
	case RFSWITCH_TX_LP:
		rc.rxPin.Set(false)
		rc.txLowPin.Set(true)
		rc.txHighPin.Set(false)
	case RFSWITCH_TX_HP:
		rc.rxPin.Set(false)
		rc.txLowPin.Set(false)
		rc.txHighPin.Set(true)
	}

	return nil
}
