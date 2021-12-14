package main

func main() {
	touchScreen, _ := initDevices()

	for {
		touch := touchScreen.ReadTouchPoint()
		if touch.Z > 0 {
			//X and Y are 16 bit with 12 bit resolution and need to be scaled for the display size
			//Z is 24 bit and is typically > 2000 for a touch
			println("touch:", touch.X, touch.Y, touch.Z)
			//Example of scaling for m5stack-core2's 320x240 display with 320x270 touch area
			println("screen:", (touch.X*320)>>16, (touch.Y*270)>>16)
		}
	}
}
