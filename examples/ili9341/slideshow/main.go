package main

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"tinygo.org/x/drivers/examples/ili9341/initdisplay"
	"tinygo.org/x/drivers/ili9341"
	"tinygo.org/x/drivers/image/jpeg"
	"tinygo.org/x/drivers/image/png"
)

var (
	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}
	red   = color.RGBA{255, 0, 0, 255}
	blue  = color.RGBA{0, 0, 255, 255}
	green = color.RGBA{0, 255, 0, 255}
)

var (
	display *ili9341.Device
)

func main() {
	err := run()
	for err != nil {
		errorMessage(err)
	}
}

func run() error {
	display = initdisplay.InitDisplay()

	width, height := display.Size()
	if width < 320 || height < 240 {
		display.SetRotation(ili9341.Rotation270)
	}

	display.FillScreen(black)

	for {
		err := drawJpeg(display)
		if err != nil {
			return err
		}
		time.Sleep(time.Second)

		err = drawPng(display)
		if err != nil {
			return err
		}
		time.Sleep(time.Second)
	}

	return nil
}

// Define the buffer required for the callback. In most cases, this setting
// should be sufficient.  For jpeg, the callback will always be called every
// 3*8*8*4 pix. png will be called every line, i.e. every width pix.
var buffer [3 * 8 * 8 * 4]uint16

func drawPng(display *ili9341.Device) error {
	p := strings.NewReader(pngImage)
	png.SetCallback(buffer[:], func(data []uint16, x, y, w, h, width, height int16) {
		err := display.DrawRGBBitmap(x, y, data[:w*h], w, h)
		if err != nil {
			errorMessage(fmt.Errorf("error drawPng: %s", err))
		}
	})

	return png.Decode(p)
}

func drawJpeg(display *ili9341.Device) error {
	p := strings.NewReader(jpegImage)
	jpeg.SetCallback(buffer[:], func(data []uint16, x, y, w, h, width, height int16) {
		err := display.DrawRGBBitmap(x, y, data[:w*h], w, h)
		if err != nil {
			errorMessage(fmt.Errorf("error drawJpeg: %s", err))
		}
	})

	return jpeg.Decode(p)
}

func errorMessage(err error) {
	for {
		fmt.Printf("%s\r\n", err.Error())
		time.Sleep(5 * time.Second)
	}
}
