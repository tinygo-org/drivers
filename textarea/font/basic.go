package font

import (
	"image/color"

	"tinygo.org/x/drivers"
)

// The color used for the background
var black = color.RGBA{0x00, 0x00, 0x00, 0x00}

const (
	firstByte     = 0 // reserved
	widthByte     = 1
	heightByte    = 2
	firstCharByte = 3
	// The number of bytes in the header
	headerLen = 4
)

type basicFont struct {
	width     int
	height    int
	firstChar int
	data      []byte
}

func newBasicFont(data []byte) basicFont {
	return basicFont{
		width:     int(data[widthByte]),
		height:    int(data[heightByte]),
		firstChar: int(data[firstCharByte]),
		data:      data[headerLen:],
	}
}

func (f basicFont) Size() (w int16, h int16) {
	return int16(f.width), int16(f.height)
}

func (f basicFont) getCharBytes(c rune) []byte {
	index := (int(c) - f.firstChar) * f.width
	if index >= len(f.data) {
		return make([]byte, f.width)
	}
	return f.data[index : index+f.width]
}

func (f basicFont) PrintChar(displayer drivers.Displayer, x, y int16, char rune, c color.RGBA) {
	dw, dh := displayer.Size()
	bytes := f.getCharBytes(char)
	// each character is rotated 90 degress clockwise, so we read bits column
	// after column to get rows
	for i := int16(0); i < int16(f.height); i++ {
		py := y + i
		for j, b := range bytes {
			px := x + int16(j)
			if px < 0 || py < 0 || px >= dw || py >= dh {
				// ignore out of bounds pixels
				continue
			}
			if (b>>i)&0x01 == 0 {
				displayer.SetPixel(px, py, black)
			} else {
				displayer.SetPixel(px, py, c)
			}
		}
	}
}

func (f basicFont) Print(displayer drivers.Displayer, x int16, y int16, str string, c color.RGBA) {
	// we cannot use the loop index that may increment by more than 1 with unicode runes
	idx := 0
	for _, char := range str {
		cx := x + int16(f.width*idx)
		f.PrintChar(displayer, cx, y, char, c)
		idx++
	}
}
