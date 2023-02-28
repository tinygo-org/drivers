package textarea

import (
	"image/color"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/textarea/font"
)

// TextArea simplifies printing text on a Displayer.
type TextArea struct {
	// Automatically wrap lines (continue them on next line)
	Wrap bool

	displayer drivers.Displayer
	ft        font.Font
	cursorX   int16
	cursorY   int16
}

// New creates a new TextArea for the given Displayer using the given font.
func New(displayer drivers.Displayer, ft font.Font) *TextArea {
	return &TextArea{
		Wrap:      false,
		displayer: displayer,
		ft:        ft,
		cursorX:   0,
		cursorY:   0,
	}
}

// Size returns the number of characters that can be printed in a row and in a column.
func (text TextArea) Size() (int16, int16) {
	w, h := text.displayer.Size()
	fw, fh := text.ft.Size()
	return w / fw, h / fh
}

// Reset resets the position of the cursor to the top-left corner.
func (text *TextArea) Reset() {
	text.cursorX = 0
	text.cursorY = 0
}

// Print prints the given string on the TextArea.
func (text *TextArea) Print(str string, c color.RGBA) {
	dw, _ := text.displayer.Size()
	fw, fh := text.ft.Size()
	for _, char := range str {
		if char == '\n' {
			text.cursorX = 0
			text.cursorY += fh
		} else if char == '\r' {
			text.cursorX = 0
		} else {
			text.ft.PrintChar(text.displayer, text.cursorX, text.cursorY, char, c)
			text.cursorX += fw
			if text.Wrap && text.cursorX >= dw {
				text.cursorX = 0
				text.cursorY += fh
			}
		}
	}
}

// Print prints the given string at the specified location on the TextArea. The location is given as a character coordinate, not a pixel coordinate.
func (text *TextArea) PrintAt(row, col int16, str string, c color.RGBA) {
	fw, fh := text.ft.Size()
	y := row * fh
	// we cannot use the loop index that may increment by more than 1 with unicode runes
	idx := 0
	for _, char := range str {
		x := (col + int16(idx)) * fw
		text.ft.PrintChar(text.displayer, x, y, char, c)
		idx++
	}
}
