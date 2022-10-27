package main

import (
        "machine"
	"image/color"
	"time"
	"strconv"
	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/st7789"
	// "tinygo.org/x/drivers/xpt2046"
	"github.com/gregoster/tinygo-drivers/xpt2046"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

func initSPI() drivers.SPI {
	machine.SPI1.Configure(machine.SPIConfig{
                Frequency: 3000000,
                Mode:      0,
        })

	return machine.SPI1

}

func initDisplay(bus drivers.SPI) (drivers.Displayer, st7789.Device) {
	
	// https://www.waveshare.com/wiki/Pico-ResTouch-LCD-2.8
        display := st7789.New(bus,
                machine.GPIO15, // TFT_RESET - reset pin
		machine.GPIO8,  // TFT_DC - Data/Command 
		machine.GPIO9,  // TFT_CS  - Chip Select
                machine.GPIO13) // TFT_LITE - Backlite pin


        display.Configure(st7789.Config{
		Rotation:   st7789.ROTATION_270,
                RowOffset:  0,
                FrameRate:  st7789.FRAMERATE_60,
                VSyncLines: st7789.MAX_VSYNC_SCANLINES,
		Width:      240,
		Height:     320,
        })
	return &display, display
}

func initTouch(bus drivers.SPI) xpt2046.Device {
	//	clk  := machine.GPIO10  // TP_CLK
        cs   := machine.GPIO16  // TP_CS
        // din  := machine.GPIO11  // MOSI
        // dout := machine.GPIO12  // MISO
        irq  := machine.GPIO17  // TP_IRQ

        touchScreen := xpt2046.New(bus, cs, irq)

        touchScreen.Configure(&xpt2046.Config{
                Precision: 10, //Maximum number of samples for a single ReadTouchPoint to improve accuracy.
        })
	return touchScreen
	
}

func main() {

	SPI := initSPI()
	
	touchScreen := initTouch(SPI)

	display, _ := initDisplay(SPI)
	width, height := display.Size()
	
        //white := color.RGBA{255, 255, 255, 255}

	// red := color.RGBA{255, 0, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	green := color.RGBA{0, 255, 0, 255}
	black := color.RGBA{0, 0, 0, 255}
	//	yellow := color.RGBA{255,255,0,255}

	tinyfont.WriteLine(display, &freemono.Regular9pt7b, 30,
		80,strconv.Itoa(int(width)) + "x" + strconv.Itoa(int(height)), green)
	
	prev := ""

	// tinyfont.WriteLine(display, &freemono.Regular9pt7b, 0, 160, "HERE!", red)
	// tinyfont.WriteLine(display, &freemono.Regular9pt7b, 0, 200, strconv.Itoa(int(machine.SPI1.GetBaudRate())), red)

	for {
		
	
                //Wait for a touch
		for !touchScreen.Touched() {
			time.Sleep(50 * time.Millisecond)
                }

		if prev != "" {
			tinyfont.WriteLine(display, &freemono.Regular9pt7b, 0,180, prev, black)
		}

		touch := touchScreen.ReadTouchPoint()

		//X and Y are 16 bit with 12 bit resolution and need to be scaled for the display size
                //Z is 24 bit and is typically > 2000 for a touch

                //Example of scaling for a 240x320 display
           
		prev = strconv.Itoa(int((touch.X*240)>>16)) + "x" + strconv.Itoa(int(touch.Y*320)>>16)
		tinyfont.WriteLine(display, &freemono.Regular9pt7b, 0, 180, prev, blue)

                //Wait for touch to end
                for touchScreen.Touched() {
                        time.Sleep(50 * time.Millisecond)
                }
	}
	
}
