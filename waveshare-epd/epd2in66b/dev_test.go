package epd2in66b

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"testing"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

type mockBus struct{}

func (m *mockBus) Tx(w, r []byte) error {
	return nil
}

func (m *mockBus) Transfer(b byte) (byte, error) {
	return 0, nil
}

func TestBufferDrawing(t *testing.T) {
	dev := New(&mockBus{})

	tinyfont.WriteLine(&dev, &freemono.Bold9pt7b, 10, 40, "Hello World!", color.RGBA{0xff, 0xff, 0xff, 0xff})

	red := color.RGBA{0xff, 0, 0, 0xff}
	black := color.RGBA{0xff, 0xff, 0xff, 0xff}
	showRect(&dev, 10, 10, 10, 10, black)
	showRect(&dev, 10, 20, 10, 10, red)

	img := toImage(&dev)
	writeImage(img)
}

func toImage(dev *Device) *image.RGBA {
	red := color.RGBA{0xff, 0, 0, 0xff}

	xMax, yMax := dev.Size()
	r := image.Rect(0, 0, int(xMax), int(yMax))
	container := image.NewRGBA(r)
	draw.Draw(container, container.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Over)

	for x := 0; x < int(xMax); x++ {
		for y := 0; y < int(yMax); y++ {

			bytePos, bitPos := pos(int16(x), int16(y), displayWidth)

			if isSet(dev.redBuffer, bytePos, bitPos) {
				container.Set(x, y, red)
			} else if isSet(dev.blackBuffer, bytePos, bitPos) {
				container.Set(x, y, color.Black)
			}
		}
	}

	return container
}

func isSet(buf []byte, bytePos, bitPos int) bool {
	return (buf[bytePos])&(0x1<<bitPos) != 0
}

func showRect(display drivers.Displayer, x int16, y int16, w int16, h int16, c color.RGBA) {
	for i := x; i < x+w; i++ {
		for j := y; j < y+h; j++ {
			display.SetPixel(i, j, c)
		}
	}
}

func writeImage(img image.Image) string {
	fn := fmt.Sprintf("%d.png", time.Now().Unix())
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		panic(err)
	}
	return fn
}
