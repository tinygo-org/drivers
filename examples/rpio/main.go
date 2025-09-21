package main

// Example program for the ST7735 display (Waveshare 1.44" LCD HAT) using the rpio driver on a Raspberry Pi.

import (
	"fmt"
	"image/color"
	"time"

	go_rpio "github.com/stianeikeland/go-rpio/v4"
	"tinygo.org/x/drivers/rpio"
	"tinygo.org/x/drivers/st7735"
)

var (
	black  = color.RGBA{0, 0, 0, 255}
	colors = [...]color.RGBA{
		{255, 0, 0, 255},   // red
		{0, 255, 0, 255},   // green
		{0, 0, 255, 255},   // blue
		{255, 255, 0, 255}, // yellow
	}
)

func main() {
	if err := go_rpio.Open(); err != nil {
		fmt.Println("Error opening GPIO:", err)
		return
	}
	defer go_rpio.Close()

	// Initialize SPI and pins
	spi := rpio.NewSPI()
	resetPin := rpio.NewPin(27)
	dcPin := rpio.NewPin(25)
	csPin := rpio.NewPin(8)
	blPin := rpio.NewPin(24)

	// Initialize display
	device := st7735.New(spi, resetPin, dcPin, csPin, blPin)
	device.Configure(st7735.Config{
		Width:        128,
		Height:       128,
		Model:        st7735.GREENTAB,
		RowOffset:    3,
		ColumnOffset: 2,
	})
	device.InvertColors(false)
	device.EnableBacklight(true)
	device.IsBGR(true) // no effect w/o rotation!
	device.SetRotation(st7735.NO_ROTATION)

	width, height := device.Size()

	// Clear display
	device.FillScreen(black)

	// Draw rectangles in a loop, clockwise rotation of colors
	pos := 0
	for {
		device.FillRectangle(0, 0, width/2, height/2, colors[(pos+0)%len(colors)])              // top left
		device.FillRectangle(0, height/2, width/2, height/2, colors[(pos+1)%len(colors)])       // bottom left
		device.FillRectangle(width/2, height/2, width/2, height/2, colors[(pos+2)%len(colors)]) // bottom right
		device.FillRectangle(width/2, 0, width/2, height/2, colors[(pos+3)%len(colors)])        // top right
		pos++
		time.Sleep(1 * time.Second)
	}
}
