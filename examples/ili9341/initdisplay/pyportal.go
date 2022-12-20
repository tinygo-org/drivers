//go:build pyportal

package initdisplay

import (
	"machine"

	"tinygo.org/x/drivers/ili9341"
)

func InitDisplay() *ili9341.Device {
	display := ili9341.NewParallel(
		machine.LCD_DATA0,
		machine.TFT_WR,
		machine.TFT_DC,
		machine.TFT_CS,
		machine.TFT_RESET,
		machine.TFT_RD,
	)

	// configure backlight
	backlight := machine.TFT_BACKLIGHT
	backlight.Configure(machine.PinConfig{machine.PinOutput})

	// configure display
	display.Configure(ili9341.Config{})

	backlight.High()

	display.SetRotation(ili9341.Rotation270)

	return display
}
