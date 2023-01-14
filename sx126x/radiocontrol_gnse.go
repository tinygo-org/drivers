//go:build gnse

/*
Generic Node Sensor Edition
RFSwitch

Disable Switch   :  PB8=OFF PA0=OFF  PA1=OFF
Enable RX        :  PB8=ON  PA0=ON   PA1=OFF
Enable TX RFO LP :  PB8=ON  PA0=ON   PA1=ON
Enable TX RFO HP :  PB8=ON  PA0=OFF  PA1=ON
*/
package sx126x

import (
	"machine"
)

// RadioControl for GNSE board.
type RadioControl struct {
	STM32RadioControl
}

func NewRadioControl() *RadioControl {
	return &RadioControl{STM32RadioControl{}}
}

// Init pins needed for controlling rx/tx
func (rc *RadioControl) Init() error {
	machine.RF_FE_CTRL1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.RF_FE_CTRL2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.RF_FE_CTRL3.Configure(machine.PinConfig{Mode: machine.PinOutput})

	return nil
}

func (rc *RadioControl) SetRfSwitchMode(mode int) error {
	switch mode {

	case RFSWITCH_TX_HP:
		machine.RF_FE_CTRL1.Set(false)
		machine.RF_FE_CTRL2.Set(true)
		machine.RF_FE_CTRL3.Set(true)

	case RFSWITCH_TX_LP:
		machine.RF_FE_CTRL1.Set(true)
		machine.RF_FE_CTRL2.Set(true)
		machine.RF_FE_CTRL3.Set(true)

	case RFSWITCH_RX:
		machine.RF_FE_CTRL1.Set(true)
		machine.RF_FE_CTRL2.Set(false)
		machine.RF_FE_CTRL3.Set(true)
	}

	return nil
}
