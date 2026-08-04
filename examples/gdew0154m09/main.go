package main

import (
	"device/esp"
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/gdew0154m09"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

func main() {
	// Hold the M5Stack CoreInk power rail on.
	machine.GPIO12.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.GPIO12.High()

	cs := machine.IO9
	dc := machine.GPIO15
	rst := machine.IO0
	busy := machine.IO4
	cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	dc.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rst.Configure(machine.PinConfig{Mode: machine.PinOutput})
	busy.Configure(machine.PinConfig{Mode: machine.PinInput})
	cs.High()
	dc.High()
	rst.High()

	// Enable and reset the ESP32 VSPI peripheral before configuring SPI3.
	esp.DPORT.SetPERIP_RST_EN_SPI3_RST(1)
	esp.DPORT.SetPERIP_CLK_EN_SPI3_CLK_EN(1)
	esp.DPORT.SetPERIP_RST_EN_SPI3_RST(0)
	if err := machine.SPI3.Configure(machine.SPIConfig{
		Frequency: gdew0154m09.Baudrate,
		SCK:       machine.IO18,
		SDO:       machine.IO23,
		SDI:       machine.NoPin,
		Mode:      gdew0154m09.SPIMode,
	}); err != nil {
		println("could not configure SPI:", err.Error())
		return
	}

	display := gdew0154m09.New(machine.SPI3, cs, dc, rst, busy)
	if err := display.Configure(gdew0154m09.DefaultConfig); err != nil {
		println("could not configure display:", err.Error())
		return
	}

	black := color.RGBA{A: 0xff}
	tinyfont.WriteLine(display, &freemono.Bold18pt7b, 15, 90, "TinyGo", black)
	for x := int16(0); x < gdew0154m09.Width; x++ {
		display.SetPixel(x, 0, black)
		display.SetPixel(x, gdew0154m09.Height-1, black)
	}
	for y := int16(0); y < gdew0154m09.Height; y++ {
		display.SetPixel(0, y, black)
		display.SetPixel(gdew0154m09.Width-1, y, black)
	}

	if err := display.Display(); err != nil {
		println("could not refresh display:", err.Error())
	}
	for {
		time.Sleep(time.Hour)
	}
}
