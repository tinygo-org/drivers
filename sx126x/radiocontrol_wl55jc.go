//go:build nucleowl55jc

/*
Nucleo WL55JC1
RFSwitch

+-----------+---------+------------+------------+
|           | FE_CTRL1 |  FE_CTRL2 |   FE_CTRL3 |
|           |   (PC4)  |   (PC5)   |     (PC3)  |
+-----------+----------+-----------+------------+
|  TX_HP    |   LOW    |   HIGH    |    HIGH    |
|  TX_LP    |   HIGH   |   HIGH    |    HIGH    |
|  RX       |   HIGH   |   LOW     |    HIGH    |
+-----------+----------+-----------+------------+
*/
package sx126x

import (
	"machine"
)

// RadioControl for Nucleo WL55JC1 board.
type RadioControl struct {
	STM32RadioControl
}

func NewRadioControl() *RadioControl {
	return &RadioControl{STM32RadioControl{}}
}

// Init pins needed for controlling rx/tx
func (rc *RadioControl) Init() error {
	machine.PC4.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.PC5.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.PC3.Configure(machine.PinConfig{Mode: machine.PinOutput})

	return nil
}

func (rc *RadioControl) SetRfSwitchMode(mode int) error {
	switch mode {

	case RFSWITCH_TX_HP:
		machine.PC4.Set(false)
		machine.PC5.Set(true)
		machine.PC3.Set(true)

	case RFSWITCH_TX_LP:
		machine.PC4.Set(true)
		machine.PC5.Set(true)
		machine.PC3.Set(true)

	case RFSWITCH_RX:
		machine.PC4.Set(true)
		machine.PC5.Set(false)
		machine.PC3.Set(true)
	}

	return nil
}
