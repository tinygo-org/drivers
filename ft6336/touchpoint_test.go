package ft6336

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"tinygo.org/x/drivers/touch"
)

func TestTouchPoint(t *testing.T) {
	tests := []struct {
		name string
		buf  []byte
		want touch.Point
	}{
		{
			name: "origin",
			buf:  []byte{1, 0x00, 0x00, 0x00, 0x00},
			want: touch.Point{X: 0, Y: 0, Z: 0xFFFFF},
		},
		{
			name: "maximum",
			buf:  []byte{1, 319 >> 8, 319 & 0xFF, 269 >> 8, 269 & 0xFF},
			want: touch.Point{X: 319 * 204, Y: 269 * 242, Z: 0xFFFFF},
		},
		{
			name: "event flag and touch ID are ignored",
			buf:  []byte{1, 0x80 | 0x01, 0x00, 0xF0 | 0x01, 0x00},
			want: touch.Point{X: 256 * 204, Y: 256 * 242, Z: 0xFFFFF},
		},
		{
			name: "two touch points",
			buf:  []byte{2, 0x00, 100, 0x00, 50},
			want: touch.Point{X: 100 * 204, Y: 50 * 242, Z: 0xFFFFF},
		},
		{
			name: "no touch",
			buf:  []byte{0, 0x00, 100, 0x00, 50},
			want: touch.Point{X: 100 * 204, Y: 50 * 242, Z: 0},
		},
		{
			name: "no touch after reset",
			buf:  []byte{255, 0xFF, 0xFF, 0xFF, 0xFF},
			want: touch.Point{X: 4095 * 204, Y: 4095 * 242, Z: 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(touchPoint(tt.buf), qt.Equals, tt.want)
		})
	}
}
