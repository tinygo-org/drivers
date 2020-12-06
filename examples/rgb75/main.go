package main

import (
	// "image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/rgb75"
)

var disp *rgb75.Device

func main() {

	disp = rgb75.New(
		machine.HUB75,
		machine.HUB75_CLK, machine.HUB75_LAT, machine.HUB75_OE,
		[6]machine.Pin{
			machine.HUB75_R1, machine.HUB75_G1, machine.HUB75_B1,
			machine.HUB75_R2, machine.HUB75_G2, machine.HUB75_B2,
		},
		[]machine.Pin{
			machine.HUB75_ADDR_A, machine.HUB75_ADDR_B, machine.HUB75_ADDR_C,
			machine.HUB75_ADDR_D, machine.HUB75_ADDR_E,
		})

	if err := disp.Configure(rgb75.Config{
		Width:      64,
		Height:     32,
		ColorDepth: 1,
	}); nil != err {
		for {
			println("error: " + err.Error())
			time.Sleep(time.Second)
		}
	}

	disp.Resume()
	for {
		// for y := int16(0); y < 32; y++ {
		// 	// 	for x := int16(0); x < 32; x++ {
		// 	disp.SetPixel(32, y, color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x00})
		// 	time.Sleep(1 * time.Second)
		// 	disp.SetPixel(32, y, color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF})
		// 	// 	}
		// }
	}
}
