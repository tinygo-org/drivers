// demo of 4-wire touchscreen as described in app note:
// http://ww1.microchip.com/downloads/en/Appnotes/doc8091.pdf
package main

import (
	"machine"
	"math"

	"tinygo.org/x/drivers/touch"
	"tinygo.org/x/drivers/touch/resistive"
)

var (
	resistiveTouch = resistive.FourWireTouchscreen{
		YP: machine.TOUCH_YD, // y+
		YM: machine.TOUCH_YU, // y-
		XP: machine.TOUCH_XR, // x+
		XM: machine.TOUCH_XL, // x-
	}
)

const (
	Xmin = 750
	Xmax = 325
	Ymin = 840
	Ymax = 240
)

func main() {

	// configure backlight
	machine.TFT_BACKLIGHT.Configure(machine.PinConfig{machine.PinOutput})

	// configure touchscreen
	machine.InitADC()
	resistiveTouch.Configure()

	last := touch.Point{}

	// loop and poll for touches, including performing debouncing
	debounce := 0
	for {

		point := resistiveTouch.GetTouchPoint()
		touch := touch.Point{}
		if point.Z > 100 {
			rawX := mapval(point.X, Xmin, Xmax, 0, 240)
			rawY := mapval(point.Y, Ymin, Ymax, 0, 320)
			touch.X = rawX
			touch.Y = rawY
			touch.Z = 1
		} else {
			touch.X = 0
			touch.Y = 0
			touch.Z = 0
		}

		if last.Z != touch.Z {
			debounce = 0
			last = touch
		} else if math.Abs(float64(touch.X-last.X)) > 4 ||
			math.Abs(float64(touch.Y-last.Y)) > 4 {
			debounce = 0
			last = touch
		} else if debounce > 1 {
			debounce = 0
			HandleTouch(last)
		} else if touch.Z > 0 {
			debounce++
		} else {
			last = touch
			debounce = 0
		}

	}
}

// based on Arduino's "map" function
func mapval(x int, inMin int, inMax int, outMin int, outMax int) int {
	return (x-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}

func HandleTouch(touch touch.Point) {
	println("touch point:", touch.X, touch.Y, touch.Z)
}
