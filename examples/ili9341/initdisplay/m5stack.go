//go:build m5stack

package initdisplay

import (
	"machine"

	"tinygo.org/x/drivers/ili9341"
)

func InitDisplay() *ili9341.Device {
	machine.SPI2.Configure(machine.SPIConfig{
		SCK:       machine.SPI0_SCK_PIN,
		SDO:       machine.SPI0_SDO_PIN,
		SDI:       machine.SPI0_SDI_PIN,
		Frequency: 40e6,
	})

	// configure backlight
	backlight := machine.LCD_BL_PIN
	backlight.Configure(machine.PinConfig{machine.PinOutput})

	display := ili9341.NewSPI(
		machine.SPI2,
		machine.LCD_DC_PIN,
		machine.LCD_SS_PIN,
		machine.LCD_RST_PIN,
	)

	// configure display
	display.Configure(ili9341.Config{
		Width:            320,
		Height:           240,
		DisplayInversion: true,
	})

	backlight.High()

	display.SetRotation(ili9341.Rotation0Mirror)

	return display
}
