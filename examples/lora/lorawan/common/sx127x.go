//go:build featherwing || lgt92

package common

import (
	"strconv"

	"machine"

	"tinygo.org/x/drivers/lora"
	"tinygo.org/x/drivers/sx127x"
)

const (
	FREQ                      = 868100000
	LORA_DEFAULT_RXTIMEOUT_MS = 1000
	LORA_DEFAULT_TXTIMEOUT_MS = 5000
)

var (
	loraRadio *sx127x.Device
)

// do sx127x setup here
func SetupLora() (lora.Radio, error) {
	rstPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	spi.Configure(machine.SPIConfig{Frequency: 500000, Mode: 0})

	loraRadio = sx127x.New(spi, rstPin)
	loraRadio.SetRadioController(sx127x.NewRadioControl(csPin, dio0Pin, dio1Pin))
	loraRadio.Reset()

	if state := loraRadio.DetectDevice(); !state {
		return nil, errRadioNotFound
	}

	return loraRadio, nil
}

func FirmwareVersion() string {
	v := loraRadio.GetVersion()
	return "sx127x v" + strconv.Itoa(int(v))
}

func Lorarx() ([]byte, error) {
	return loraRadio.Rx(LORA_DEFAULT_RXTIMEOUT_MS)
}
