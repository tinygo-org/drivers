package main

import (
	"image/color"
)

// trail contains a queue of pixels, a coordinate for the head of the queue to
// move into, and a width dimension defining the horizontal range of head.
type trail struct {
	pix []pixel
	pos point
	dim int
	fac float32
}

func newTrail(xSpan int, ySpeed float32) *trail {
	return &trail{
		pix: []pixel{
			{point: noPoint, color: color.RGBA{R: 0x0, G: 0x0, B: 0x0, A: 0xF}},
			{point: noPoint, color: color.RGBA{R: 0x1, G: 0x0, B: 0x0, A: 0xF}},
			{point: noPoint, color: color.RGBA{R: 0x3, G: 0x0, B: 0x0, A: 0xF}},
			{point: noPoint, color: color.RGBA{R: 0x7, G: 0x0, B: 0x0, A: 0xF}},
			{point: noPoint, color: color.RGBA{R: 0xF, G: 0x0, B: 0x0, A: 0xF}},
			{point: noPoint, color: color.RGBA{R: 0xF, G: 0x0, B: 0x0, A: 0xF}},
			{point: noPoint, color: color.RGBA{R: 0xF, G: 0x0, B: 0x1, A: 0xF}},
			{point: noPoint, color: color.RGBA{R: 0x7, G: 0x0, B: 0x3, A: 0xF}},
			{point: noPoint, color: color.RGBA{R: 0x3, G: 0x0, B: 0x7, A: 0xF}},
			{point: noPoint, color: color.RGBA{R: 0x1, G: 0x0, B: 0xF, A: 0xF}},
			{point: noPoint, color: color.RGBA{R: 0x0, G: 0xF, B: 0x0, A: 0xF}},
			{point: noPoint, color: color.RGBA{R: 0x0, G: 0xF, B: 0x0, A: 0xF}},
		},
		pos: top(0, xSpan),
		dim: xSpan,
		fac: ySpeed,
	}
}

// push enqueues each of the given points to the receiver trail in the order
// they were provided.
func (t *trail) push(ps ...point) {
	for _, p := range ps {
		for i, px := range t.pix[1:] {
			t.pix[i].point = px.point
		}
		t.pix[len(t.pix)-1].point = p
	}
}

// inc pushes the trail head onto the pixel queue and then increments head,
// returning its new coordinates.
func (t *trail) inc() point {
	// add the current position to the list
	t.push(t.pos)
	// update the current position
	t.pos = t.pos.next(t.fac)
	return t.pos
}

// wrap resets the trail head to the top of the screen.
func (t *trail) wrap() {
	t.pos = top(0, t.dim)
}
