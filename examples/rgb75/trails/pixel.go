package main

import (
	"image/color"
)

// pixel represents an individual RGB color with given point coordinates.
type pixel struct {
	point point
	color color.RGBA
}
