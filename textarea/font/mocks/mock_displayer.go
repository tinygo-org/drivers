package mocks

import "image/color"

// MockDisplayer is only by tests to easily check if data is properly sent to a Displayer. It uses a grid where each pixel is represented with a byte and has a 0 or 1 value.
type MockDisplayer struct {
	width, height int16
	pixels        [][]byte
}

// NewMockDisplayer creates a new MockDisplayer with the given dimensions in pixels.
func NewMockDisplayer(width int16, height int16) *MockDisplayer {
	pixels := make([][]byte, height)
	for i := range pixels {
		pixels[i] = make([]byte, width)
	}
	return &MockDisplayer{
		width:  width,
		height: height,
		pixels: pixels,
	}
}

// SetPixel implements the Displayer method.
func (d *MockDisplayer) SetPixel(x int16, y int16, c color.RGBA) {
	var b byte
	if c.R != 0 || c.G != 0 || c.B != 0 {
		b = 1
	} else {
		b = 0
	}
	d.pixels[y][x] = b
}

// GetPixels returns the pixels grid.
func (d *MockDisplayer) GetPixels() [][]byte {
	return d.pixels
}

// Display implements the Displayer method.
func (d *MockDisplayer) Display() error {
	return nil
}

// Size implements the Displayer method.
func (d *MockDisplayer) Size() (x, y int16) {
	return d.width, d.height
}
