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
	resistiveTouch = new(resistive.FourWire)
)

const (
	Xmin = 750
	Xmax = 325
	Ymin = 840
	Ymax = 240
)

func main() {

	// configure touchscreen
	machine.InitADC()
	resistiveTouch.Configure(&resistive.FourWireConfig{
		YP: machine.TOUCH_YD, // y+
		YM: machine.TOUCH_YU, // y-
		XP: machine.TOUCH_XR, // x+
		XM: machine.TOUCH_XL, // x-
	})

	last := touch.Point{}

	// loop and poll for touches, including performing debouncing
	debounce := 0
	for {

		point := resistiveTouch.ReadTouchPoint()
		touch := touch.Point{}
		if point.Z>>6 > 100 {
			touch.X = mapval(point.X>>6, Xmin, Xmax, 0, 240)
			touch.Y = mapval(point.Y>>6, Ymin, Ymax, 0, 320)
			touch.Z = point.Z >> 6 / 100
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
