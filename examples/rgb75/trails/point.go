package main

import (
	"math/rand"
)

// point represents a 2-dimensional point in space.
type point struct{ x, y float32 }

// noPoint represents a point that exists outside of any screen's coordinate
// space.
var noPoint = point{x: -256.0, y: -256.0}

// pos returns the x and y components of the receiver, rounded to the nearest
// int16 integer. Note that x and y may be negative.
func (p point) pos() (x, y int16) {
	round := func(f float32) int16 {
		// naive rounding (half-away)
		if f < 0 {
			f -= 0.5
		} else {
			f += 0.5
		}
		return int16(f)
	}
	return round(p.x), round(p.y)
}

// next returns a new point whose x and y components are equal to the receiver's
// x and y components incremented by random deltas.
// The x component delta is a random value in the interval (-1,1).
// The y component delta is a random value in the interval [0,1), multiplied by
// the given factor speed.
func (p point) next(speed float32) point {
	dx, dy := rand.Float32(), rand.Float32()*speed
	if rand.Intn(2) == 0 {
		return point{x: p.x - dx, y: p.y + dy}
	}
	return point{x: p.x + dx, y: p.y + dy}
}

// top returns a new point positioned at the top of a screen.
// The x component is a random integer in the interval [xMin,xMax).
// The y component is always 0.
func top(xMin, xMax int) point {
	if xMin > xMax {
		xMin, xMax = xMax, xMin
	}
	return point{
		x: float32(int32(xMin) + rand.Int31n(int32(xMax-xMin))),
		y: 0.0,
	}
}
