package image

import "image/color"

type Image struct {
	width   int16
	height  int16
	xMirror bool
	yMirror bool
	color   []uint32
}

func (i *Image) Size() (width, height int16) {
	return i.width, i.height
}

func (i *Image) Mirror(x, y bool) {
	i.xMirror = x
	i.yMirror = y
}

func (i *Image) ColorAt(x, y int16) (c color.RGBA) {
	if n := y*i.width + x; int(n) < len(i.color) {
		u := i.color[n]
		c = color.RGBA{
			R: uint8((u >> 24) & 0xFF),
			G: uint8((u >> 16) & 0xFF),
			B: uint8((u >> 8) & 0xFF),
			A: uint8(u & 0xFF),
		}
	}
	return
}
