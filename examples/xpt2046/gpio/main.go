package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/xpt2046"
)

func main() {

	clk := machine.GPIO0
	cs := machine.GPIO1
	din := machine.GPIO2
	dout := machine.GPIO3
	irq := machine.GPIO4

	touchScreen := xpt2046.New(clk, cs, din, dout, irq)

	touchScreen.Configure(&xpt2046.Config{
		Precision: 10, //Maximum number of samples for a single ReadTouchPoint to improve accuracy.
	})

	for {

		//Wait for a touch
		for !touchScreen.Touched() {
			time.Sleep(50 * time.Millisecond)
		}

		touch := touchScreen.ReadTouchPoint()
		//X and Y are 16 bit with 12 bit resolution and need to be scaled for the display size
		//Z is 24 bit and is typically > 2000 for a touch
		println("touch:", touch.X, touch.Y, touch.Z)
		//Example of scaling for a 240x320 display
		println("screen:", (touch.X*240)>>16, (touch.Y*320)>>16)

		//Wait for touch to end
		for touchScreen.Touched() {
			time.Sleep(50 * time.Millisecond)
		}

	}
}
