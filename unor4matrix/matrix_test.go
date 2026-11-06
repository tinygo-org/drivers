package unor4matrix

import (
	"image/color"
	"testing"

	"tinygo.org/x/drivers"
)

var _ drivers.Displayer = (*Device)(nil)

func TestClearDisplay(t *testing.T) {
	var d Device
	d.LoadFrame([3]uint32{0xffffffff, 0xffffffff, 0xffffffff})
	d.ClearDisplay()
	if d.frame.pixels != ([framebufferSize]byte{}) {
		t.Fatalf("pixels = %v", d.frame.pixels)
	}
}

func TestLoadFrame(t *testing.T) {
	allOn := [12]byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	}
	tests := []struct {
		name  string
		words [3]uint32
		want  [12]byte
	}{
		{
			name: "blank replaces existing pixels",
		},
		{
			name:  "all pixels",
			words: [3]uint32{0xffffffff, 0xffffffff, 0xffffffff},
			want:  allOn,
		},
		{
			name:  "first pixel",
			words: [3]uint32{0x80000000, 0, 0},
			want:  [12]byte{0: 0x01},
		},
		{
			name:  "last bit of first word",
			words: [3]uint32{0x00000001, 0, 0},
			want:  [12]byte{3: 0x80},
		},
		{
			name:  "first bit of second word",
			words: [3]uint32{0, 0x80000000, 0},
			want:  [12]byte{4: 0x01},
		},
		{
			name:  "last bit of second word",
			words: [3]uint32{0, 0x00000001, 0},
			want:  [12]byte{7: 0x80},
		},
		{
			name:  "first bit of third word",
			words: [3]uint32{0, 0, 0x80000000},
			want:  [12]byte{8: 0x01},
		},
		{
			name:  "last pixel",
			words: [3]uint32{0, 0, 0x00000001},
			want:  [12]byte{11: 0x80},
		},
		{
			name:  "bit order within a byte",
			words: [3]uint32{0xa1000000, 0, 0},
			want:  [12]byte{0: 0x85},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Device
			d.frame.pixels = allOn

			d.LoadFrame(tt.words)
			if d.frame.pixels != tt.want {
				t.Fatalf("drawing buffer = %v, want %v", d.frame.pixels, tt.want)
			}
		})
	}
}

func TestSizeAndColors(t *testing.T) {
	var d Device
	if w, h := d.Size(); w != 12 || h != 8 {
		t.Fatalf("size %d,%d", w, h)
	}
	for _, c := range []color.RGBA{{R: 1}, {G: 1}, {B: 1}, {R: 255, A: 0}} {
		d.ClearDisplay()
		d.SetPixel(11, 7, c)
		if d.frame.pixels != ([framebufferSize]byte{11: 0x80}) {
			t.Fatalf("color %v should be on", c)
		}
	}
	d.SetPixel(11, 7, color.RGBA{A: 255})
	if d.frame.pixels != ([framebufferSize]byte{}) {
		t.Fatal("black should be off")
	}
	before := d.frame
	for _, xy := range [][2]int16{{-1, 0}, {0, -1}, {12, 0}, {0, 8}} {
		d.SetPixel(xy[0], xy[1], color.RGBA{R: 255})
	}
	if d.frame != before {
		t.Fatal("out-of-bounds changed drawing buffer")
	}
}

func TestMapping(t *testing.T) {
	if len(ledMapping) != 96 {
		t.Fatal("mapping length", len(ledMapping))
	}
	seen := map[pinPair]bool{}
	for i, pair := range ledMapping {
		if pair.anode >= 11 || pair.cathode >= 11 || pair.anode == pair.cathode || seen[pair] {
			t.Fatalf("invalid/duplicate pair %d: %v", i, pair)
		}
		seen[pair] = true
	}
	if ledMapping[0] != (pinPair{7, 3}) || ledMapping[1] != (pinPair{3, 7}) || ledMapping[95] != (pinPair{9, 4}) {
		t.Fatal("known source landmarks differ")
	}
}
