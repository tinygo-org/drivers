//go:build !stm32wlx && sx126x

package common

import (
	"machine"

	"tinygo.org/x/drivers/lora"
	"tinygo.org/x/drivers/sx126x"
)

const (
	FREQ                      = 868100000
	LORA_DEFAULT_RXTIMEOUT_MS = 1000
	LORA_DEFAULT_TXTIMEOUT_MS = 5000
)

var (
	loraRadio *sx126x.Device
)

var (
	spi                        = machine.SPI0
	nssPin, busyPin, dio1Pin   = machine.GP17, machine.GP10, machine.GP11
	rxPin, txLowPin, txHighPin = machine.GP13, machine.GP12, machine.GP12
)

func newRadioControl() sx126x.RadioController {
	return sx126x.NewRadioControl(nssPin, busyPin, dio1Pin, rxPin, txLowPin, txHighPin)
}

// do sx126x setup here
func SetupLora() (lora.Radio, error) {
	loraRadio = sx126x.New(spi)
	loraRadio.SetDeviceType(sx126x.DEVICE_TYPE_SX1262)

	// Create radio controller for target
	loraRadio.SetRadioController(newRadioControl())

	if state := loraRadio.DetectDevice(); !state {
		return nil, errRadioNotFound
	}

	return loraRadio, nil
}

func FirmwareVersion() string {
	return "sx126x"
}

func Lorarx() ([]byte, error) {
	return loraRadio.Rx(LORA_DEFAULT_RXTIMEOUT_MS)
}
