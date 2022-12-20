//go:build feather_m0 || feather_m4 || feather_m4_can || feather_nrf52840 || feather_nrf52840_sense || feather_stm32f405 || feather_rp2040

package initdisplay

import (
	"machine"

	"tinygo.org/x/drivers/ili9341"
)

func InitDisplay() *ili9341.Device {
	machine.D5.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.D6.Configure(machine.PinConfig{Mode: machine.PinOutput})

	machine.SPI0.Configure(machine.SPIConfig{
		SCK:       machine.SPI0_SCK_PIN,
		SDO:       machine.SPI0_SDO_PIN,
		SDI:       machine.SPI0_SDI_PIN,
		Frequency: 40000000,
	})

	// configure backlight
	backlight := machine.D9
	backlight.Configure(machine.PinConfig{machine.PinOutput})

	display := ili9341.NewSPI(
		machine.SPI0,
		machine.D10, // LCD_DC,
		machine.D11, // LCD_SS_PIN,
		machine.D12, // LCD_RESET,
	)

	// configure display
	display.Configure(ili9341.Config{})

	backlight.High()

	display.SetRotation(ili9341.Rotation270)

	return display
}
