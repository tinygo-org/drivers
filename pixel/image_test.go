package pixel_test

import (
	"image/color"
	"testing"

	"tinygo.org/x/drivers/pixel"
)

func TestImageRGB565BE(t *testing.T) {
	image := pixel.NewImage[pixel.RGB565BE](5, 3)
	if width, height := image.Size(); width != 5 && height != 3 {
		t.Errorf("image.Size(): expected 5, 3 but got %d, %d", width, height)
	}
	for _, c := range []color.RGBA{
		{R: 0xff, A: 0xff},
		{G: 0xff, A: 0xff},
		{B: 0xff, A: 0xff},
		{R: 0x10, A: 0xff},
		{G: 0x10, A: 0xff},
		{B: 0x10, A: 0xff},
	} {
		image.Set(4, 2, pixel.NewColor[pixel.RGB565BE](c.R, c.G, c.B))
		c2 := image.Get(4, 2).RGBA()
		if c2 != c {
			t.Errorf("failed to roundtrip color: expected %v but got %v", c, c2)
		}
	}
}

func TestImageRGB444BE(t *testing.T) {
	image := pixel.NewImage[pixel.RGB444BE](5, 3)
	if width, height := image.Size(); width != 5 && height != 3 {
		t.Errorf("image.Size(): expected 5, 3 but got %d, %d", width, height)
	}
	for _, c := range []color.RGBA{
		{R: 0xff, A: 0xff},
		{G: 0xff, A: 0xff},
		{B: 0xff, A: 0xff},
		{R: 0x11, A: 0xff},
		{G: 0x11, A: 0xff},
		{B: 0x11, A: 0xff},
	} {
		encoded := pixel.NewColor[pixel.RGB444BE](c.R, c.G, c.B)
		image.Set(0, 0, encoded)
		image.Set(0, 1, encoded)
		encoded2 := image.Get(0, 0)
		encoded3 := image.Get(0, 1)
		if encoded != encoded2 {
			t.Errorf("failed to roundtrip color %v: expected %d but got %d", c, encoded, encoded2)
		}
		if encoded != encoded3 {
			t.Errorf("failed to roundtrip color %v: expected %d but got %d", c, encoded, encoded3)
		}
		c2 := encoded2.RGBA()
		if c2 != c {
			t.Errorf("failed to roundtrip color: expected %v but got %v", c, c2)
		}
		c3 := encoded3.RGBA()
		if c3 != c {
			t.Errorf("failed to roundtrip color: expected %v but got %v", c, c3)
		}
	}
}
