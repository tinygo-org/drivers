package main

import (
	"image/color"
	"machine"
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/adafruit4650"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	dev := adafruit4650.New(machine.I2C0)

	err := dev.Configure()
	if err != nil {
		panic(err)
	}

	drawPlus(&dev)
	drawHelloWorld(&dev)

	err = dev.Display()
	if err != nil {
		panic(err)
	}
}

func drawPlus(d drivers.Displayer) {
	for i := int16(0); i < 128; i++ {
		d.SetPixel(i, 32, color.RGBA{R: 1})
	}
	for i := int16(0); i < 64; i++ {
		d.SetPixel(64, i, color.RGBA{R: 1})
	}
}

func drawHelloWorld(d drivers.Displayer) {
	tinyfont.WriteLine(d, &freemono.Regular9pt7b, 0, 32, "Hello World!", color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
}
