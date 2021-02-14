package main

import (
	"math"
	"math/rand"
)

const (
	dimensionMax = Dimension(32767)
	dimensionMin = Dimension(0)
	positionMax  = Position(int32(velocityMax) * int32(dimensionMax))
	positionMin  = Position(0)
	velocityMax  = Velocity(256)
	velocityMin  = -velocityMax
	velocityMax2 = velocityMax * velocityMax // max²
)

// Dimension represents one component of a discrete pixel's 2D coordinates
type Dimension uint16 // (physical space)

// Position converts the receiver d in physical space to a component of a
// Particle coordinate in logical space.
func (d Dimension) Position() Position {
	return Position(int(d) * int(velocityMax))
}

// Position represents one component of a discrete Particle's 2D coordinates
type Position int32 // (logical space)

// Dimension converts the receiver p in logical space to a component of a real
// Pixel coordinate in physical space.
func (p Position) Dimension() Dimension {
	return Dimension(int(p) / int(velocityMax))
}

// Move returns the receiver Position p adjusted by given Velocity v.
func (p Position) Move(v Velocity) Position {
	return Position(int(p) + int(v))
}

// Velocity represents one component of a discrete Particle's 2D velocity
type Velocity int32 // (logical space)

// Reverse returns the receiver Velocity v in the opposite direction and scaled
// by a given elasticity.
func (v Velocity) Reverse(elasticity int) Velocity {
	return Velocity(int(-v) * elasticity / int(velocityMax))
}

// Abs returns the absolute value of the receiver Velocity v.
func (v Velocity) Abs() Velocity {
	if v < 0 {
		return -v
	}
	return v
}

// Particle represents an object moving through space.
//
// The space through which a Particle moves is referred to in documentation as
// "logical space", since that space is much larger than the "physical space"
// used to describe physical Pixel coordinates; these added logical coordinates
// exist "in-between" Pixels, and allow for smoother movement in the absence of
// floating-point coordinates.
//
// Particles in logical space are always eventually projected onto physical
// space when displaying them with a Pixel.
type Particle struct {
	index  int
	x, y   Position
	vx, vy Velocity
}

// Particles represents all Particle objects in the Field's 2D space.
type Particles []Particle

// ParticleMove defines a callback used to notify callers when a Particle moves.
type ParticleMove func(f *Field, p *Particle, x, y Position)

// MakeParticles returns a new Particle buffer of given Field f and count n.
// Each Particle is initially positioned in the first unoccupied Pixel on the
// Field.
func MakeParticles(f *Field, n int) Particles {
	if n >= 0 {
		particle := make(Particles, n)
		for i := range particle {
			x := Dimension(i) / f.height
			y := Dimension(i) % f.height
			particle[i].index = i
			particle[i].SetPosition(f, x.Position(), y.Position())
		}
		return particle
	}
	return nil
}

// Accelerate applies the current acceleration due to gravity to the velocity of
// receiver p, with a slight perturbation epsilon.
// This only changes the Particle velocity; it does not affect its Position.
func (p *Particle) Accelerate(x, y, z, epsilon int) {
	// apply random perturbation to the values read from accelerometer.
	// do not use MakeVelocity, as it will prematurely clip the x, y components.
	p.vx += Velocity(x + rand.Intn(epsilon))
	p.vy += Velocity(y + rand.Intn(epsilon))
	// clip the resulting vector to maximum velocity
	v2 := p.vx*p.vx + p.vy*p.vy
	if v2 > velocityMax2 { // implies v > velocityMax in some direction
		v := math.Sqrt(float64(v2))
		p.vx = Velocity(int(float64(velocityMax*p.vx) / v))
		p.vy = Velocity(int(float64(velocityMax*p.vy) / v))
	}
}

// SetPosition sets the (x, y) Position coordinates of the receiver Particle p
// in logical space, and updates the Obstacle coordinates of the given Field f
// in physical space.
func (p *Particle) SetPosition(f *Field, x, y Position) {
	if nil != f.handleMove {
		f.handleMove(f, p, x, y)
	}
	f.obstacle.Clr(p.x.Dimension(), p.y.Dimension())
	p.x, p.y = x, y
	f.obstacle.Set(p.x.Dimension(), p.y.Dimension())
}

// Move attempts to change the logical Position coordinates of the receiver
// Particle p based on its current velocity, or reverses velocity if the change
// of Position would collide with an Obstacle.
func (p *Particle) Move(f *Field) {

	// first, compute destination Position based on current Velocity
	x := p.x.Move(p.vx)
	y := p.y.Move(p.vy)

	// next, verify we are moving within Field boundaries
	if x < 0 {
		p.vx = p.vx.Reverse(f.elasticity)
		x = 0
	} else if x > f.xMax {
		p.vx = p.vx.Reverse(f.elasticity)
		x = f.xMax
	}
	if y < 0 {
		p.vy = p.vy.Reverse(f.elasticity)
		y = 0
	} else if y > f.yMax {
		p.vy = p.vy.Reverse(f.elasticity)
		y = f.yMax
	}

	// then, determine if we are moving into a new real Pixel in physical space
	if dp := f.PixelIndex(p.x, p.y) - f.PixelIndex(x, y); 0 != dp {
		// check if the destination Pixel contains an Obstacle
		if f.obstacle.Get(x.Dimension(), y.Dimension()) {
			if dp < 0 {
				dp = -dp // absolute value of index difference
			}
			// determine which direction the Obstacle exists
			switch dp {
			case 1: // obstructed by 1 pixel to the left or right
				p.vx = p.vx.Reverse(f.elasticity)
				x = p.x

			case int(f.width): // obstructed by 1 pixel to the top or bottom
				p.vy = p.vy.Reverse(f.elasticity)
				y = p.y

			default: // obstructed by 1 pixel in a diagonal direction
				if p.vx.Abs() >= p.vy.Abs() {
					if !f.obstacle.Get(x.Dimension(), p.y.Dimension()) {
						p.vy = p.vy.Reverse(f.elasticity)
						y = p.y
					} else if !f.obstacle.Get(p.x.Dimension(), y.Dimension()) {
						p.vx = p.vx.Reverse(f.elasticity)
						x = p.x
					} else {
						p.vx = p.vx.Reverse(f.elasticity)
						p.vy = p.vy.Reverse(f.elasticity)
						x, y = p.x, p.y
					}
				} else {
					if !f.obstacle.Get(p.x.Dimension(), y.Dimension()) {
						p.vx = p.vx.Reverse(f.elasticity)
						x = p.x
					} else if !f.obstacle.Get(x.Dimension(), p.y.Dimension()) {
						p.vy = p.vy.Reverse(f.elasticity)
						y = p.y
					} else {
						p.vx = p.vx.Reverse(f.elasticity)
						p.vy = p.vy.Reverse(f.elasticity)
						x, y = p.x, p.y
					}
				}
			}
		}
	}

	// update coordinates of both Particle (logical space) and Obstacle (physical
	// space) using new (x, y) Positions.
	p.SetPosition(f, x, y)
}
