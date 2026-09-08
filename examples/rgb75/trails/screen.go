package main

import (
	"tinygo.org/x/drivers/rgb75"
)

// screen associates an embedded rgb75.Device with a list of trails to animate.
type screen struct {
	*rgb75.Device
	trail []*trail
}

// contains returns true if and only if the given point exists within the
// receiver's screen dimensions.
func (s *screen) contains(p point) bool {
	width, height := s.Size()
	aboveMin := p.x >= 0 && p.y >= 0
	belowMax := int16(p.x+0.5) < width && int16(p.y+0.5) < height
	return aboveMin && belowMax
}
