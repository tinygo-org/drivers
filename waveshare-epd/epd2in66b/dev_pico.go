//go:build pico

package epd2in66b

import (
	"machine"
)

// DefaultConfig contains the default config for the https://www.waveshare.com/wiki/Pico-ePaper-2.66 module
var DefaultConfig = Config{
	DataPin:       machine.GP8,
	ChipSelectPin: machine.GP9,
	ResetPin:      machine.GP12,
	BusyPin:       machine.GP13,
}

// NewPicoModule allocates a new device backed by the https://www.waveshare.com/wiki/Pico-ePaper-2.66 module
// This will also configure the SPI1 bus and configure the device with the DefaultConfig
func NewPicoModule() (Device, error) {
	spi := machine.SPI1

	if err := spi.Configure(machine.SPIConfig{
		Frequency: Baudrate,
	}); err != nil {
		return Device{}, err
	}

	dev := New(spi)
	if err := dev.Configure(DefaultConfig); err != nil {
		return dev, err
	}

	return dev, nil
}
