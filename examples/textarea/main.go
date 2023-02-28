package main

import (
	"image/color"
	"time"

	"tinygo.org/x/drivers/examples/ili9341/initdisplay"
	"tinygo.org/x/drivers/ili9341"
	"tinygo.org/x/drivers/textarea"
	"tinygo.org/x/drivers/textarea/font"
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
	display = initdisplay.InitDisplay()

	text := textarea.New(display, font.NewFont6x8())

	black := color.RGBA{0x00, 0x00, 0x00, 0xFF}
	white := color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}

	text.Print("Initializing...\n", white)
	text.Print("GPS...", white)
	time.Sleep(time.Second * 1)
	text.Print("ok\n", white)

	text.Print("GPS...", white)
	time.Sleep(time.Second * 1)
	text.Print("ok\n", white)

	text.Print("WiFi...", white)
	time.Sleep(time.Second * 1)
	text.Print("ok\n", white)

	display.FillScreen(black)
	text.Reset()

	_, h := text.Size()
	text.PrintAt(h-1, 0, "Idle", white)

	for {
		time.Sleep(time.Hour)
	}
}
