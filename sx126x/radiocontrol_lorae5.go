//go:build lorae5

/*
LoRa-E5 module ONLY transmits through RFO_HP:

Receive: PA4=1, PA5=0
Transmit(high output power, SMPS mode): PA4=0, PA5=1
*/

package sx126x

import (
	"machine"
)

// RadioControl for LoRa-E5 board.
type RadioControl struct {
	STM32RadioControl
}

func NewRadioControl() *RadioControl {
	return &RadioControl{STM32RadioControl: STM32RadioControl{}}
}

// Init pins needed for controlling rx/tx
func (rc *RadioControl) Init() error {
	machine.PA4.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.PB5.Configure(machine.PinConfig{Mode: machine.PinOutput})

	return nil
}

func (rc *RadioControl) SetRfSwitchMode(mode int) error {
	switch mode {

	case RFSWITCH_RX:
		machine.PA4.Set(true)
		machine.PB5.Set(false)
	case RFSWITCH_TX_LP:
		return errLowPowerTxNotSupported
	case RFSWITCH_TX_HP:
		machine.PA4.Set(false)
		machine.PB5.Set(true)

	}
	return nil
}
