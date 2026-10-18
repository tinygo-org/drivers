package uc8151

import (
	"image/color"
	"testing"
)

type fakeOutputPin struct{}

func (p *fakeOutputPin) Set(level bool) {}

type fakeInputPin struct {
	high bool
}

func (p *fakeInputPin) Get() bool {
	return p.high
}

type fakeSPI struct{}

func (s *fakeSPI) Transfer(value byte) (byte, error) {
	return 0, nil
}

func (s *fakeSPI) Tx(write, read []byte) error {
	return nil
}

func newTestDevice(t *testing.T) *Device {
	t.Helper()
	// The busy pin is high when the display is idle
	busy := &fakeInputPin{high: true}
	device := New(&fakeSPI{}, &fakeOutputPin{}, &fakeOutputPin{}, &fakeOutputPin{}, busy)
	device.Configure(Config{})
	return &device
}

func pixelSet(device *Device, x, y int16) bool {
	return device.buffer[x/8+y*(device.width/8)]&(0x80>>uint8(x%8)) != 0
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
			if got := pixelSet(device, x, y); got != want {
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
			if pixelSet(device, x, y) {
				t.Fatalf("pixel %d,%d = true, want false", x, y)
			}
		}
	}
}
