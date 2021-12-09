package main

import (
	"image/color"
	"machine"
	"math/rand"

	"tinygo.org/x/drivers/ssd1289"
)

func main() {

	//The SSD1289 is configured in 16 bit parallel mode  and requires 16 GPIOs
	//The Pin bus is the most flexible but ineffecient method it switches
	//individual pins on and off. If you are able to use consecutive pins
	//consider creating a more efficient bus implementation that uses
	//your microcontrollers built in "ports"
	//see rp2040bus.go for an example for the rapsberry pi pico
	bus := ssd1289.NewPinBus([16]machine.Pin{
		machine.GP4,  //DB0
		machine.GP5,  //DB1
		machine.GP6,  //DB2
		machine.GP7,  //DB3
		machine.GP8,  //DB4
		machine.GP9,  //DB5
		machine.GP10, //DB6
		machine.GP11, //DB7
		machine.GP12, //DB8
		machine.GP13, //DB9
		machine.GP14, //DB10
		machine.GP15, //DB11
		machine.GP16, //DB12
		machine.GP17, //DB13
		machine.GP18, //DB14
		machine.GP19, //DB15
	})

	//Control pins for the SSD1289
	rs := machine.GP0
	wr := machine.GP1
	cs := machine.GP2
	rst := machine.GP3

	display := ssd1289.New(rs, wr, cs, rst, bus)

	display.Configure() //Sends intialization sequence to SSD1289.
	//!! After configure the display will contain random data and needs to be cleared

	background := color.RGBA{0, 0, 0, 255} //Black
	display.FillDisplay(background)        //Clears the display to the given color

	for {
		//Draw random filled coloured rectangles
		x := int16(rand.Intn(120))
		w := int16(rand.Intn(120))

		y := int16(rand.Intn(160))
		h := int16(rand.Intn(160))

		r := uint8(rand.Intn(255))
		g := uint8(rand.Intn(255))
		b := uint8(rand.Intn(255))

		c := color.RGBA{r, g, b, 255}

		display.FillRect(x, y, w, h, c) //Fills the given rectangle the rest of the display is unaffected.
	}
}
