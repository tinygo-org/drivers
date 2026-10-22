package ssd1306

import (
	"image/color"
	"testing"
)

type fakeI2C struct{}

func (b *fakeI2C) Tx(addr uint16, write, read []byte) error {
	return nil
}

func newTestDevice(t *testing.T) *Device {
	t.Helper()
	device := NewI2C(&fakeI2C{})
	device.Configure(Config{Width: 128, Height: 64})
	return device
}

func TestFillRectangleAtBottomRightCorner(t *testing.T) {
	device := newTestDevice(t)
	width, height := device.Size()
	device.FillRectangle(0, 0, width, height, color.RGBA{255, 255, 255, 255})

	// A black line ending at the bottom-right corner must not clear the whole display
	if err := device.FillRectangle(width-28, height-1, 28, 1, color.RGBA{0, 0, 0, 255}); err != nil {
		t.Fatalf("FillRectangle() error = %v", err)
	}
	for x := int16(0); x < width; x++ {
		for y := int16(0); y < height; y++ {
			want := !(x >= width-28 && y == height-1)
			if got := device.GetPixel(x, y); got != want {
				t.Fatalf("pixel %d,%d = %v, want %v", x, y, got, want)
			}
		}
	}
}

func TestFillRectangleFullScreenBlack(t *testing.T) {
	device := newTestDevice(t)
	width, height := device.Size()
	device.FillRectangle(0, 0, width, height, color.RGBA{255, 255, 255, 255})

	if err := device.FillRectangle(0, 0, width, height, color.RGBA{0, 0, 0, 255}); err != nil {
		t.Fatalf("FillRectangle() error = %v", err)
	}
	for x := int16(0); x < width; x++ {
		for y := int16(0); y < height; y++ {
			if device.GetPixel(x, y) {
				t.Fatalf("pixel %d,%d = true, want false", x, y)
			}
		}
	}
}
