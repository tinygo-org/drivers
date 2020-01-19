package main

import (
	"image/color"
	"machine"
	"math"

	"tinygo.org/x/drivers/ili9341"
	"tinygo.org/x/drivers/touch"
	"tinygo.org/x/drivers/touch/resistive"
)

var (
	resistiveTouch = &resistive.FourWire{}

	display = ili9341.NewParallel(
		machine.LCD_DATA0,
		machine.TFT_WR,
		machine.TFT_DC,
		machine.TFT_CS,
		machine.TFT_RESET,
		machine.TFT_RD,
	)

	white   = color.RGBA{255, 255, 255, 255}
	black   = color.RGBA{0, 0, 0, 255}
	red     = color.RGBA{255, 0, 0, 255}
	green   = color.RGBA{0, 255, 0, 255}
	blue    = color.RGBA{0, 0, 255, 255}
	magenta = color.RGBA{255, 0, 255, 255}
	yellow  = color.RGBA{255, 255, 0, 255}
	cyan    = color.RGBA{0, 255, 255, 255}

	oldColor     color.RGBA
	currentColor color.RGBA
)

const (
	penRadius = 3
	boxSize   = 30

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
	resistiveTouch.Configure(&resistive.FourWireConfig{
		YP: machine.TOUCH_YD,
		YM: machine.TOUCH_YU,
		XP: machine.TOUCH_XR,
		XM: machine.TOUCH_XL,
	})

	// configure display
	display.Configure(ili9341.Config{})

	// fill the background and activate the backlight
	width, height := display.Size()
	display.FillRectangle(0, 0, width, height, black)
	machine.TFT_BACKLIGHT.High()

	// make color selection boxes
	display.FillRectangle(0, 0, boxSize, boxSize, red)
	display.FillRectangle(boxSize, 0, boxSize, boxSize, yellow)
	display.FillRectangle(boxSize*2, 0, boxSize, boxSize, green)
	display.FillRectangle(boxSize*3, 0, boxSize, boxSize, cyan)
	display.FillRectangle(boxSize*4, 0, boxSize, boxSize, blue)
	display.FillRectangle(boxSize*5, 0, boxSize, boxSize, magenta)
	display.FillRectangle(boxSize*6, 0, boxSize, boxSize, black)
	display.FillRectangle(boxSize*7, 0, boxSize, boxSize, white)

	// set the initial color to red and draw a box to highlight it
	oldColor = red
	currentColor = red
	display.DrawRectangle(0, 0, boxSize, boxSize, white)

	last := touch.Point{}

	// loop and poll for touches, including performing debouncing
	debounce := 0
	for {

		point := resistiveTouch.ReadTouchPoint()
		touch := touch.Point{}
		if point.Z>>6 > 100 {
			rawX := mapval(point.X>>6, Xmin, Xmax, 0, 240)
			rawY := mapval(point.Y>>6, Ymin, Ymax, 0, 320)
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

	if int16(touch.Y) < boxSize {
		oldColor = currentColor
		x := int16(touch.X)
		switch {
		case x < boxSize:
			currentColor = red
		case x < boxSize*2:
			currentColor = yellow
		case x < boxSize*3:
			currentColor = green
		case x < boxSize*4:
			currentColor = cyan
		case x < boxSize*5:
			currentColor = blue
		case x < boxSize*6:
			currentColor = magenta
		case x < boxSize*7:
			currentColor = black
		case x < boxSize*8:
			currentColor = white
		}

		if oldColor == currentColor {
			return
		}

		display.DrawRectangle((x/boxSize)*boxSize, 0, boxSize, boxSize, white)
		switch oldColor {
		case red:
			x = 0
		case yellow:
			x = boxSize
		case green:
			x = boxSize * 2
		case cyan:
			x = boxSize * 3
		case blue:
			x = boxSize * 4
		case magenta:
			x = boxSize * 5
		case black:
			x = boxSize * 6
		case white:
			x = boxSize * 7
		}
		display.FillRectangle(int16(x), 0, boxSize, boxSize, oldColor)

	}

	if (int16(touch.Y) - penRadius) > boxSize {
		display.FillRectangle(
			int16(touch.X), int16(touch.Y), penRadius*2, penRadius*2, currentColor)
	}
}
