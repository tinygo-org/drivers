package ws2812

import (
	"image/color"
	"testing"
)

func TestApplyBrightness(t *testing.T) {
	for _, tc := range []struct {
		c          color.RGBA
		brightness uint8
		r, g, b    uint8
	}{
		{color.RGBA{R: 0xff, G: 0xff, B: 0xff}, 255, 0xff, 0xff, 0xff},
		{color.RGBA{R: 0xff, G: 0x80, B: 0x01}, 255, 0xff, 0x80, 0x01},
		{color.RGBA{R: 0xff, G: 0xff, B: 0xff}, 0, 0, 0, 0},
		{color.RGBA{R: 0xff, G: 0x80, B: 0x00}, 127, 0x7f, 0x40, 0},
		{color.RGBA{R: 0x00, G: 0xff, B: 0x80}, 63, 0, 0x3f, 0x20},
	} {
		r, g, b := applyBrightness(tc.c, tc.brightness)
		if r != tc.r || g != tc.g || b != tc.b {
			t.Errorf("applyBrightness(%v, %d) = %#x %#x %#x, want %#x %#x %#x",
				tc.c, tc.brightness, r, g, b, tc.r, tc.g, tc.b)
		}
	}
}

func TestApplyBrightnessRange(t *testing.T) {
	scale := func(v, brightness uint8) uint8 {
		r, _, _ := applyBrightness(color.RGBA{R: v}, brightness)
		return r
	}
	for v := range 256 {
		if got := scale(uint8(v), 255); got != uint8(v) {
			t.Errorf("value %#x at brightness 255 = %#x, want %#x", v, got, v)
		}
		if got := scale(uint8(v), 0); got != 0 {
			t.Errorf("value %#x at brightness 0 = %#x, want 0", v, got)
		}
		prev := uint8(0)
		for b := range 256 {
			got := scale(uint8(v), uint8(b))
			if got > uint8(v) {
				t.Fatalf("value %#x at brightness %d = %#x, more than the input", v, b, got)
			}
			if got < prev {
				t.Fatalf("value %#x: brightness %d gives %#x, less than %#x at %d", v, b, got, prev, b-1)
			}
			prev = got
		}
	}
}
