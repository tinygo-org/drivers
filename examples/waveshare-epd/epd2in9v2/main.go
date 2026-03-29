package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers/waveshare-epd/epd2in9v2"
)

var display epd2in9v2.Device

func main() {
	machine.SPI0.Configure(machine.SPIConfig{
		Frequency: 12000000,
		SCK:       machine.EPD_SCK_PIN,
		SDO:       machine.EPD_SDO_PIN,
	})

	display = epd2in9v2.New(
		machine.SPI0,
		machine.EPD_CS_PIN,
		machine.EPD_DC_PIN,
		machine.EPD_RESET_PIN,
		machine.EPD_BUSY_PIN,
	)
	display.Configure(epd2in9v2.Config{
		Rotation: epd2in9v2.ROTATION_270,
		Speed:    epd2in9v2.SPEED_DEFAULT,
		Blocking: true,
	})

	black := color.RGBA{0, 0, 0, 255}
	white := color.RGBA{255, 255, 255, 255}

	// --- Step 1: clear to white ---
	println("epd2in9v2: clearing display")
	display.ClearBuffer()
	display.Display()
	time.Sleep(2 * time.Second)

	// --- Step 2: full refresh checkerboard ---
	println("epd2in9v2: drawing checkerboard (full refresh)")
	w, h := display.Size()
	for i := int16(0); i < w/8; i++ {
		for j := int16(0); j < h/8; j++ {
			if (i+j)%2 == 0 {
				fillRect(i*8, j*8, 8, 8, black)
			}
		}
	}
	display.Display()
	time.Sleep(2 * time.Second)

	// --- Step 3: fast refresh - draw border and diagonal cross ---
	println("epd2in9v2: switching to fast refresh")
	display.SetSpeed(epd2in9v2.SPEED_FAST)
	display.ClearBuffer()

	for x := int16(0); x < w; x++ {
		display.SetPixel(x, 0, black)
		display.SetPixel(x, h-1, black)
	}
	for y := int16(0); y < h; y++ {
		display.SetPixel(0, y, black)
		display.SetPixel(w-1, y, black)
	}
	for i := int16(0); i < w && i < h; i++ {
		display.SetPixel(i, i*h/w, black)
		display.SetPixel(w-1-i, i*h/w, black)
	}

	display.Display()
	time.Sleep(2 * time.Second)

	// --- Step 4: partial refresh counter ---
	println("epd2in9v2: partial refresh demo")
	display.SetSpeed(epd2in9v2.SPEED_DEFAULT)
	display.ClearBuffer()

	println("epd2in9v2: setting base image")
	display.DisplayWithBase()

	for count := 0; count < 10; count++ {
		cx := int16(120)
		cy := int16(50)
		fillRect(cx, cy, 60, 20, white)

		digit := int16(count % 10)
		drawDigit(cx+22, cy+2, digit, black)

		display.DisplayPartial()
		time.Sleep(500 * time.Millisecond)
	}
	time.Sleep(2 * time.Second)

	// --- Step 5: sleep ---
	println("epd2in9v2: entering deep sleep")
	display.ClearBuffer()
	display.Display()
	display.Sleep()
	println("epd2in9v2: done, you can remove power")
}

func fillRect(x, y, w, h int16, c color.RGBA) {
	for i := x; i < x+w; i++ {
		for j := y; j < y+h; j++ {
			display.SetPixel(i, j, c)
		}
	}
}

// drawDigit draws a simple 3x5-pixel-block digit (each block 4x3 px) at position (x,y).
func drawDigit(x, y, digit int16, c color.RGBA) {
	segments := [10][5]uint8{
		{0x7, 0x5, 0x5, 0x5, 0x7}, // 0
		{0x2, 0x2, 0x2, 0x2, 0x2}, // 1
		{0x7, 0x1, 0x7, 0x4, 0x7}, // 2
		{0x7, 0x1, 0x7, 0x1, 0x7}, // 3
		{0x5, 0x5, 0x7, 0x1, 0x1}, // 4
		{0x7, 0x4, 0x7, 0x1, 0x7}, // 5
		{0x7, 0x4, 0x7, 0x5, 0x7}, // 6
		{0x7, 0x1, 0x1, 0x1, 0x1}, // 7
		{0x7, 0x5, 0x7, 0x5, 0x7}, // 8
		{0x7, 0x5, 0x7, 0x1, 0x7}, // 9
	}
	if digit < 0 || digit > 9 {
		return
	}
	for row := int16(0); row < 5; row++ {
		for col := int16(0); col < 3; col++ {
			if segments[digit][row]&(0x4>>uint(col)) != 0 {
				fillRect(x+col*4, y+row*3, 4, 3, c)
			}
		}
	}
}
