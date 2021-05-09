// This example shows how to use 128x64 display over I2C
// Tested on Seeeduino XIAO Expansion Board https://wiki.seeedstudio.com/Seeeduino-XIAO-Expansion-Board/
//
// According to manual, I2C address of the display is 0x78, but that's 8-bit address.
// TinyGo operates on 7-bit addresses and respective 7-bit address would be 0x3C, which we use below.
//
// To learn more about different types of I2C addresses, please see following page
// https://www.totalphase.com/support/articles/200349176-7-bit-8-bit-and-10-bit-I2C-Slave-Addressing

package main

import (
	"machine"

	"image/color"
	"time"

	"tinygo.org/x/drivers/ssd1306"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	display := ssd1306.NewI2C(machine.I2C0)
	display.Configure(ssd1306.Config{
		Address: 0x3C,
		Width:   128,
		Height:  64,
	})

	display.ClearDisplay()

	x := int16(0)
	y := int16(0)
	deltaX := int16(1)
	deltaY := int16(1)
	for {
		pixel := display.GetPixel(x, y)
		c := color.RGBA{255, 255, 255, 255}
		if pixel {
			c = color.RGBA{0, 0, 0, 255}
		}
		display.SetPixel(x, y, c)
		display.Display()

		x += deltaX
		y += deltaY

		if x == 0 || x == 127 {
			deltaX = -deltaX
		}

		if y == 0 || y == 63 {
			deltaY = -deltaY
		}
		time.Sleep(1 * time.Millisecond)
	}
}
