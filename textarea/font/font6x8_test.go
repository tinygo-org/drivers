package font

import (
	"testing"
)

func TestNewFont6x8(t *testing.T) {
	font := NewFont6x8()
	w, h := font.Size()
	if w != 6 || h != 8 {
		t.Errorf("Font6x8.Size() = (%d,%d), want (%d,%d)", w, h, 6, 8)
	}
}
