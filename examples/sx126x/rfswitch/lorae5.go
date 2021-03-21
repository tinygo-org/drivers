//go:build lorae5
// +build lorae5

package radio

/*
/!\ LoRa-E5 module ONLY transmits through RFO_HP:

	Receive: PA4=1, PA5=0
	Transmit(high output power, SMPS mode): PA4=0, PA5=1

*/

import (
	"errors"
	"machine"

	"tinygo.org/x/drivers/sx126x"
)

type CustomSwitch struct {
}

func (s CustomSwitch) InitRFSwitch() {
	machine.PA4.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.PB5.Configure(machine.PinConfig{Mode: machine.PinOutput})
}

func (s CustomSwitch) SetRfSwitchMode(mode int) error {
	switch mode {

	case sx126x.RFSWITCH_RX:
		machine.PA4.Set(true)
		machine.PB5.Set(false)
	case sx126x.RFSWITCH_TX_LP:
		errors.New("RFSWITCH_TX_LP not supported ")
	case sx126x.RFSWITCH_TX_HP:
		machine.PA4.Set(false)
		machine.PB5.Set(true)

	}
	return nil
}
