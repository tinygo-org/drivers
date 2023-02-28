package font

import (
	"image/color"

	"tinygo.org/x/drivers"
)

type Font interface {
	Size() (w int16, h int16)
	PrintChar(displayer drivers.Displayer, x, y int16, char rune, c color.RGBA)
	Print(displayer drivers.Displayer, x int16, y int16, str string, c color.RGBA)
}

func NewFont(data []byte) Font {
	return newBasicFont(data)
}
