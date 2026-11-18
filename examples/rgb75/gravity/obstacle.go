package main

// Obstacle represents objects with fixed coordinates in space a Particle may
// not occupy.
type Obstacle uint8 // (physical space) stored as bitmap for RAM efficiency

// Obstacles represents all Obstacle objects in the Field's 2D space.
type Obstacles struct {
	wc Dimension
	ob []Obstacle
}

// MakeObstacles allocates and returns a buffer for storing coordinate indices
// of every Obstacle in the Field.
func MakeObstacles(w, h Dimension) Obstacles {
	wc := (w + 7) / 8
	return Obstacles{
		wc: wc,
		ob: make([]Obstacle, int(wc)*int(h)),
	}
}

// Index returns the Obstacle bitmap buffer index for the given real Pixel
// coordinates from physical space, and whether or not that index is valid.
func (o *Obstacles) Index(x, y Dimension) (int, bool) {
	n := int(y*o.wc + x/8)
	return n, n >= 0 && n < len(o.ob)
}

// Set defines the existence of an Obstacle at the given real Pixel coordinates
// from physical space.
func (o *Obstacles) Set(x, y Dimension) {
	if n, ok := o.Index(x, y); ok {
		o.ob[n] |= Obstacle(int(0x80) >> int(x&7))
	}
}

// Clr removes any existence of an Obstacle at the given real Pixel coordinates
// from physical space.
func (o *Obstacles) Clr(x, y Dimension) {
	if n, ok := o.Index(x, y); ok {
		o.ob[n] &= Obstacle(int(0x7F7F) >> int(x&7))
	}
}

// Get returns true if and only if an Obstacle exists at the given real Pixel
// coordinates from physical space.
func (o *Obstacles) Get(x, y Dimension) bool {
	if n, ok := o.Index(x, y); ok {
		return 0 != (int(o.ob[n]) & (int(0x80) >> int(x&7)))
	}
	return true // outside of Field, assume it is an obstacle
}
