//go:build gnse
// +build gnse

/*
Generic Node Sensor Edition
RFSwitch

Disable Switch   :  PB8=OFF PA0=OFF  PA1=OFF
Enable RX        :  PB8=ON  PA0=ON   PA1=OFF
Enable TX RFO LP :  PB8=ON  PA0=ON   PA1=ON
Enable TX RFO HP :  PB8=ON  PA0=OFF  PA1=ON
*/
package rfswitch

import (
	"machine"

	"tinygo.org/x/drivers/sx126x"
)

type CustomSwitch struct {
}

var (
	rfstate int
)

func (s CustomSwitch) InitRFSwitch() {
	machine.RF_FE_CTRL1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.RF_FE_CTRL2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.RF_FE_CTRL3.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rfstate = -1 //Unknown
}

func (s CustomSwitch) SetRfSwitchMode(mode int) error {
	if mode == rfstate {
		return nil
	}

	switch mode {

	case sx126x.RFSWITCH_TX_HP:
		machine.RF_FE_CTRL1.Set(false)
		machine.RF_FE_CTRL2.Set(true)
		machine.RF_FE_CTRL3.Set(true)

	case sx126x.RFSWITCH_TX_LP:
		machine.RF_FE_CTRL1.Set(true)
		machine.RF_FE_CTRL2.Set(true)
		machine.RF_FE_CTRL3.Set(true)

	case sx126x.RFSWITCH_RX:
		machine.RF_FE_CTRL1.Set(true)
		machine.RF_FE_CTRL2.Set(false)
		machine.RF_FE_CTRL3.Set(true)
	}

	rfstate = mode

	return nil
}
