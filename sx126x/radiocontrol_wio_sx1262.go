//go:build wio_sx1262

package sx126x

import (
	"machine"
)

// NewRadioControlWio creates RadioControl configured for wio1262 module.
// The board has RF_SW pin which controls RX/TX, but providing it is not required
// if SetDio2AsRfSwitchCtrl is set to true (use machine.NoPin instead).
// Additionally, the board requires following settings on initalization:
/*
	radio.SetDio3AsTcxoCtrl(sx126x.SX126X_DIO3_OUTPUT_1_8, 5*time.Millisecond)
 	radio.SetRegulatorMode(sx126x.SX126X_REGULATOR_DC_DC)
	radio.SetDeviceType(sx126x.DEVICE_TYPE_SX1262)
	radio.Calibrate(sx126x.SX126X_CALIBRATE_ALL)
*/
func NewRadioControlWio(nssPin, busyPin, dio1Pin, rfSw machine.Pin) *RadioControl {
	return &RadioControl{
		nssPin:    nssPin,
		busyPin:   busyPin,
		dio1Pin:   dio1Pin,
		rxPin:     rfSw,
		txLowPin:  machine.NoPin,
		txHighPin: machine.NoPin,
	}
}
