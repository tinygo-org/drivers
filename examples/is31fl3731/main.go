package main

import (
	"time"

	"machine"

	"tinygo.org/x/drivers/is31fl3731"
)

// I2CAddress -- address of led matrix
var I2CAddress uint8 = is31fl3731.I2C_ADDRESS_74

func main() {
	bus := machine.I2C0
	err := bus.Configure(machine.I2CConfig{})
	if err != nil {
		println("could not configure I2C:", err)
		return
	}

	// Create driver for Adafruit 15x7 CharliePlex LED Matrix FeatherWing
	// (CharlieWing): https://www.adafruit.com/product/3163
	ledMatrix := is31fl3731.NewAdafruitCharlieWing15x7(bus, I2CAddress)

	err = ledMatrix.Configure()
	if err != nil {
		println("could not configure is31fl3731 driver:", err)
		return
	}

	// Fill the whole matrix on the frame #0 (visible by default)
	ledMatrix.Fill(is31fl3731.FRAME_0, uint8(3))

	// Draw couple pixels on the frame #1 (not visible yet)
	ledMatrix.DrawPixelXY(is31fl3731.FRAME_1, uint8(0), uint8(0), uint8(10))
	ledMatrix.DrawPixelXY(is31fl3731.FRAME_1, uint8(14), uint8(6), uint8(10))

	// There are 8 frames available, it's a good idea to draw on an invisible
	// frame and then switch to that frame to reduce flickering. Switch between
	// frame #0 and #1 in a loop to show animation:
	for {
		println("show frame #0...")
		ledMatrix.SetActiveFrame(is31fl3731.FRAME_0)
		time.Sleep(time.Second * 3)

		println("show frame #1...")
		ledMatrix.SetActiveFrame(is31fl3731.FRAME_1)
		time.Sleep(time.Second * 3)
	}
}
