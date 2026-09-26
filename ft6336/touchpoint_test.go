package ft6336

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/touch"
)

func TestTouchPoint(t *testing.T) {
	tests := []struct {
		name          string
		buf           []byte
		width, height int
		want          touch.Point
	}{
		{
			name:  "origin",
			buf:   []byte{1, 0x00, 0x00, 0x00, 0x00},
			width: 320, height: 270,
			want: touch.Point{X: 0, Y: 0, Z: 0xFFFFF},
		},
		{
			name:  "maximum",
			buf:   []byte{1, 319 >> 8, 319 & 0xFF, 269 >> 8, 269 & 0xFF},
			width: 320, height: 270,
			want: touch.Point{X: 0xFFFF, Y: 0xFFFF, Z: 0xFFFFF},
		},
		{
			name:  "event flag and touch ID are ignored",
			buf:   []byte{1, 0x80 | 0x01, 0x00, 0xF0 | 0x01, 0x00},
			width: 320, height: 270,
			want: touch.Point{X: 256 * 0xFFFF / 319, Y: 256 * 0xFFFF / 269, Z: 0xFFFFF},
		},
		{
			name:  "two touch points",
			buf:   []byte{2, 0x00, 100, 0x00, 50},
			width: 320, height: 270,
			want: touch.Point{X: 100 * 0xFFFF / 319, Y: 50 * 0xFFFF / 269, Z: 0xFFFFF},
		},
		{
			name:  "no touch",
			buf:   []byte{0, 0x00, 100, 0x00, 50},
			width: 320, height: 270,
			want: touch.Point{X: 100 * 0xFFFF / 319, Y: 50 * 0xFFFF / 269, Z: 0},
		},
		{
			name:  "no touch after reset",
			buf:   []byte{255, 0xFF, 0xFF, 0xFF, 0xFF},
			width: 320, height: 270,
			want: touch.Point{X: 0xFFFF, Y: 0xFFFF, Z: 0},
		},
		{
			name:  "maximum 240x320",
			buf:   []byte{1, 239 >> 8, 239 & 0xFF, 319 >> 8, 319 & 0xFF},
			width: 240, height: 320,
			want: touch.Point{X: 0xFFFF, Y: 0xFFFF, Z: 0xFFFFF},
		},
		{
			name:  "center 240x320",
			buf:   []byte{1, 0x00, 108, 0x00, 167},
			width: 240, height: 320,
			want: touch.Point{X: 108 * 0xFFFF / 239, Y: 167 * 0xFFFF / 319, Z: 0xFFFFF},
		},
		{
			name:  "larger than the panel",
			buf:   []byte{1, 320 >> 8, 320 & 0xFF, 270 >> 8, 270 & 0xFF},
			width: 320, height: 270,
			want: touch.Point{X: 0xFFFF, Y: 0xFFFF, Z: 0xFFFFF},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(touchPoint(tt.buf, tt.width, tt.height), qt.Equals, tt.want)
		})
	}
}

func TestPanelSize(t *testing.T) {
	tests := []struct {
		name                  string
		width, height         int
		wantWidth, wantHeight int
	}{
		{name: "defaults", width: 0, height: 0, wantWidth: 320, wantHeight: 270},
		{name: "width only", width: 240, height: 0, wantWidth: 240, wantHeight: 270},
		{name: "height only", width: 0, height: 320, wantWidth: 320, wantHeight: 320},
		{name: "both", width: 240, height: 320, wantWidth: 240, wantHeight: 320},
		{name: "negative", width: -1, height: -1, wantWidth: 320, wantHeight: 270},
		{name: "one", width: 1, height: 1, wantWidth: 320, wantHeight: 270},
		{name: "two", width: 2, height: 2, wantWidth: 2, wantHeight: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := qt.New(t)
			width, height := panelSize(tt.width, tt.height)
			c.Assert(width, qt.Equals, tt.wantWidth)
			c.Assert(height, qt.Equals, tt.wantHeight)
		})
	}
}
