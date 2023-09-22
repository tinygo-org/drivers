package adafruit4650

import (
	"bytes"
	_ "embed"
	"encoding/hex"
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

//go:embed expected_hello_world.png
var expectedHelloWorld []byte

// mockBus mocks a fake i2c device adafruit4650 display.
// The memory layout assumes that clients set up the device in a particular way and always send complete
// pages to the device buffer.
type mockBus struct {
	img           draw.Image
	line          int
	addr          uint8
	currentPage   int
	currentColumn int
}

func (m *mockBus) Tx(addr uint16, w, r []byte) error {
	if addr != uint16(m.addr) {
		panic("unexpected address")
	}
	if r != nil {
		panic("mock does not support reads")
	}

	if w[0] == 0x00 {
		if w[1]&0xf0 == 0xb0 {
			m.currentPage = int(w[1] & 0x0f)

			lo := w[2] & 0x0f
			hi := w[2] & 0x07
			m.currentColumn = int(hi<<4 | lo)
		}
		return nil
	}
	if w[0] != 0x40 {
		panic("unexpected first byte: " + hex.EncodeToString(w[0:1]))
	}

	return m.writeRAM(w[1:])
}

func newMock() *mockBus {

	m := image.NewRGBA(image.Rect(0, 0, width, height))
	return &mockBus{img: m, addr: DefaultAddress, currentPage: -1, currentColumn: -1}
}

func (m *mockBus) writeRAM(data []byte) error {

	// RAM layout
	//    *-----> y
	//    |
	//   x|     col0  col1  ... col63
	//    v  p0  a0    b0         ..
	//           a1    b1         ..
	//           ..    ..         ..
	//           a7    b7         ..
	//       p1  a0    b0
	//           a1    b1
	//

	fmt.Printf("writing page %d\n", m.currentPage)
	// assuming entire pages will be written
	for x := 0; x < 8; x++ {
		for y := 0; y < height; y++ {

			col := data[y]

			c := color.Black
			if col&(1<<x) != 0 {
				c = color.White
			}

			m.img.Set(x+m.currentPage*8, height-y-1, c)
		}
	}

	return nil
}

func (m *mockBus) toImage() *image.RGBA {

	container := image.NewRGBA(m.img.Bounds().Inset(-1))
	draw.Draw(container, container.Bounds(), image.NewUniform(color.RGBA{G: 255, A: 255}), image.Point{}, draw.Over)
	draw.Draw(container, m.img.Bounds(), m.img, image.Point{}, draw.Over)
	return container
}

func TestDevice_Display(t *testing.T) {

	bus := newMock()
	dev := New(bus)

	dev.Configure()

	drawPlus(&dev)
	drawHellowWorld(&dev)

	//when
	dev.Display()

	//then
	actual := bus.toImage()

	expected, err := png.Decode(bytes.NewReader(expectedHelloWorld))
	if err != nil {
		panic(err)
	}

	assertEqualImages(t, actual, expected)
}

func drawPlus(d drivers.Displayer) {
	for i := int16(0); i < 128; i++ {
		d.SetPixel(i, 32, color.RGBA{R: 1})
	}
	for i := int16(0); i < 64; i++ {
		d.SetPixel(64, i, color.RGBA{R: 1})
	}
}

func drawHellowWorld(d drivers.Displayer) {
	tinyfont.WriteLine(d, &freemono.Regular9pt7b, 0, 32, "Hello World!", color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
}

func assertEqualImages(t testing.TB, actual, expected image.Image) {

	if actual.Bounds().Dx() != expected.Bounds().Dx() || actual.Bounds().Dy() != expected.Bounds().Dy() {
		f := writeImage(actual)
		t.Fatalf("differing size: was %v, expected %v, saved actual to %s", actual.Bounds(), expected.Bounds(), f)
	}

	bb := expected.Bounds()
	for x := bb.Min.X; x < bb.Max.X; x++ {
		for y := bb.Min.Y; y < bb.Max.Y; y++ {
			actualBB := actual.Bounds()
			if actual.At(x+actualBB.Min.X, y+actualBB.Min.Y) != expected.At(x, y) {
				f := writeImage(actual)
				t.Fatalf("different pixel at %d/%d: %v != %v, saved actual at %s", x, y, actual.At(x, y), expected.At(x, y), f)
			}
		}
	}
}

func writeImage(img image.Image) string {

	fn := fmt.Sprintf("%d.png", time.Now().Unix())
	f, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, 0644)
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
