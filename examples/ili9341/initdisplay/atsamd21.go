//go:build atsamd21

package initdisplay

import (
	"machine"

	"tinygo.org/x/drivers/ili9341"
)

func InitDisplay() *ili9341.Device {
	machine.SPI0.Configure(machine.SPIConfig{
		SCK:       machine.SPI0_SCK_PIN,
		SDO:       machine.SPI0_SDO_PIN,
		SDI:       machine.SPI0_SDI_PIN,
		Frequency: 24000000,
	})

	// configure backlight
	backlight := machine.D3
	backlight.Configure(machine.PinConfig{machine.PinOutput})

	display := ili9341.NewSPI(
		machine.SPI0,
		machine.D0,
		machine.D1,
		machine.D2,
	)

	// configure display
	display.Configure(ili9341.Config{})

	backlight.High()

	display.SetRotation(ili9341.Rotation270)

	return display
}
