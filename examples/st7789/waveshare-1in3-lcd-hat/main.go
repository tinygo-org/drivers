// Demo for the Waveshare 1.3inch LCD HAT on a Waveshare RP2040-PiZero.
//
// The HAT has a 240x240 ST7789VW panel, a 5 way joystick and three keys.
// The demo shows a joystick cross and three key boxes. Each part fills with
// colour while it is pressed. The joystick moves a dot in the top area of the
// screen. A press on the joystick centre moves the dot back to the middle.
// Product wiki: https://www.waveshare.com/wiki/1.3inch_LCD_HAT
//
// Board support.
//
// The pin numbers below are only correct for the RP2040-PiZero. Any other
// board needs its own mapping.
//
//	Waveshare RP2040-PiZero          this demo
//	Raspberry Pi Pico and Pico W     untested, needs its own pin mapping
//	                                 because it has no 40 pin header
//	Other RP2040 boards              untested, needs its own pin mapping
//
// The RP2040-PiZero has a Raspberry Pi 40 pin header. Waveshare exchanges
// GPIO10 and GPIO11 on the header so that RP2040 SPI1 SCK and TX align with
// the Raspberry Pi SCLK and MOSI positions.
// Schematic: https://files.waveshare.com/wiki/RP2040-PiZero/RP2040-PiZero.pdf
//
// Header to RP2040 GPIO for the HAT signals:
//
//	pin 13 RST    GPIO27
//	pin 18 BL     GPIO24
//	pin 19 MOSI   GPIO11 (SPI1 SDO)
//	pin 22 DC     GPIO25
//	pin 23 SCLK   GPIO10 (SPI1 SCK)
//	pin 24 CS     GPIO8
//	pin 29 LEFT   GPIO15
//	pin 31 UP     GPIO6
//	pin 33 PRESS  GPIO13
//	pin 35 DOWN   GPIO19
//	pin 36 KEY3   GPIO16
//	pin 37 RIGHT  GPIO26
//	pin 38 KEY2   GPIO20
//	pin 40 KEY1   GPIO21
//
// The joystick and the keys connect to ground when pressed. The HAT has no
// pull up resistors, so the demo enables the RP2040 internal pull ups.
//
// Build and flash:
//
//	tinygo flash -target=waveshare-rp2040-pizero ./examples/st7789/waveshare-1in3-lcd-hat
package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/st7789"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/proggy"
)

const (
	lcdRST = machine.GPIO27
	lcdDC  = machine.GPIO25
	lcdCS  = machine.GPIO8
	lcdBL  = machine.GPIO24

	pollInterval = 10 * time.Millisecond
	cursorSize   = 6
	cursorStep   = 2
)

var (
	black  = color.RGBA{0, 0, 0, 255}
	white  = color.RGBA{255, 255, 255, 255}
	grey   = color.RGBA{96, 96, 96, 255}
	red    = color.RGBA{255, 0, 0, 255}
	green  = color.RGBA{0, 255, 0, 255}
	yellow = color.RGBA{255, 255, 0, 255}
)

type box struct {
	x, y, w, h int16
}

type input struct {
	pin    machine.Pin
	label  string
	box    box
	colour color.RGBA
	down   bool
}

func (in *input) pressed() bool {
	return !in.pin.Get()
}

// The top area holds the cursor. The bottom area holds the indicators.
var area = box{0, 0, 240, 128}

var (
	up    = &input{pin: machine.GPIO6, label: "U", box: box{48, 138, 26, 26}, colour: green}
	down  = &input{pin: machine.GPIO19, label: "D", box: box{48, 202, 26, 26}, colour: green}
	left  = &input{pin: machine.GPIO15, label: "L", box: box{16, 170, 26, 26}, colour: green}
	right = &input{pin: machine.GPIO26, label: "R", box: box{80, 170, 26, 26}, colour: green}
	press = &input{pin: machine.GPIO13, label: "P", box: box{48, 170, 26, 26}, colour: yellow}
	key1  = &input{pin: machine.GPIO21, label: "KEY1", box: box{140, 138, 84, 26}, colour: red}
	key2  = &input{pin: machine.GPIO20, label: "KEY2", box: box{140, 170, 84, 26}, colour: red}
	key3  = &input{pin: machine.GPIO16, label: "KEY3", box: box{140, 202, 84, 26}, colour: red}

	inputs = []*input{up, down, left, right, press, key1, key2, key3}
)

func main() {
	// The HAT has no MISO line. SDI must be NoPin, because the zero value is
	// GPIO0 and SPI1 cannot use GPIO0.
	err := machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 40_000_000,
		SCK:       machine.SPI1_SCK_PIN,
		SDO:       machine.SPI1_SDO_PIN,
		SDI:       machine.NoPin,
		Mode:      0,
	})
	if err != nil {
		for {
			println("SPI1 configure:", err.Error())
			time.Sleep(time.Second)
		}
	}

	display := st7789.New(machine.SPI1, lcdRST, lcdDC, lcdCS, lcdBL)
	display.Configure(st7789.Config{
		Width:    240,
		Height:   240,
		Rotation: drivers.Rotation90,
	})

	for _, in := range inputs {
		in.pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}

	display.FillScreen(black)
	drawFrame(&display, area, grey)
	for _, in := range inputs {
		drawInput(&display, in)
	}

	cx, cy := centre()
	drawCursor(&display, cx, cy, white)

	for {
		for _, in := range inputs {
			if d := in.pressed(); d != in.down {
				in.down = d
				drawInput(&display, in)
			}
		}

		nx, ny := cx, cy
		if up.down {
			ny -= cursorStep
		}
		if down.down {
			ny += cursorStep
		}
		if left.down {
			nx -= cursorStep
		}
		if right.down {
			nx += cursorStep
		}
		if press.down {
			nx, ny = centre()
		}
		nx = clamp(nx, area.x+1, area.x+area.w-1-cursorSize)
		ny = clamp(ny, area.y+1, area.y+area.h-1-cursorSize)

		if nx != cx || ny != cy {
			drawCursor(&display, cx, cy, black)
			cx, cy = nx, ny
			drawCursor(&display, cx, cy, white)
		}

		time.Sleep(pollInterval)
	}
}

func centre() (int16, int16) {
	return area.x + (area.w-cursorSize)/2, area.y + (area.h-cursorSize)/2
}

func clamp(v, lo, hi int16) int16 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func drawCursor(d *st7789.Device, x, y int16, c color.RGBA) {
	d.FillRectangle(x, y, cursorSize, cursorSize, c)
}

func drawFrame(d *st7789.Device, b box, c color.RGBA) {
	d.DrawFastHLine(b.x, b.x+b.w-1, b.y, c)
	d.DrawFastHLine(b.x, b.x+b.w-1, b.y+b.h-1, c)
	d.DrawFastVLine(b.x, b.y, b.y+b.h-1, c)
	d.DrawFastVLine(b.x+b.w-1, b.y, b.y+b.h-1, c)
}

func drawInput(d *st7789.Device, in *input) {
	b := in.box
	fill, text := black, in.colour
	if in.down {
		fill, text = in.colour, black
	}
	d.FillRectangle(b.x+1, b.y+1, b.w-2, b.h-2, fill)
	drawFrame(d, b, in.colour)

	_, w := tinyfont.LineWidth(&proggy.TinySZ8pt7b, in.label)
	tx := b.x + (b.w-int16(w))/2
	ty := b.y + b.h/2 + 4
	tinyfont.WriteLine(d, &proggy.TinySZ8pt7b, tx, ty, in.label, text)
}
