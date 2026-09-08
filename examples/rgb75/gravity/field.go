package main

// Field contains the entire coordinate spaces for both physical Pixels and the
// logical Particles that move through them, providing methods for interacting
// with and projecting objects from one space to the other.
type Field struct {
	width, height Dimension
	xMax, yMax    Position

	accelScale int
	elasticity int

	obstacle Obstacles
	particle Particles

	handleMove ParticleMove
}

// Default values for members of type FieldConfig.
const (
	DefaultAccelScale = 1
	DefaultElasticity = 128
)

// FieldConfig defines a Field's general configuration options.
type FieldConfig struct {
	Width, Height int
	NumParticles  int
	AccelScale    int
	Elasticity    int
	HandleMove    ParticleMove
}

// NewField allocates and returns a new Field. The Configure method must be
// called before the object can be used.
func NewField() *Field {
	return &Field{}
}

// Configure sets a Field's general configuration obtions and allocates storage
// for the various buffers used to hold obstacles, particles, and pixels.
func (f *Field) Configure(config FieldConfig) error {

	// set physical field dimensions (number of pixels)
	f.width = Dimension(config.Width)
	f.height = Dimension(config.Height)

	// set logical field dimensions (particle space)
	f.xMax = f.width.Position() - 1
	f.yMax = f.height.Position() - 1

	// use default accelerometer scaling unless a non-zero value was given
	f.accelScale = DefaultAccelScale
	if config.AccelScale > 0 {
		f.accelScale = config.AccelScale
	}

	// since 0 is a valid elasticity, but in reality not what the user probably
	// intended, we need some other way to determine if the default value should
	// be used.
	// instead, check to see if ALL optional config parameters are set to 0, and
	// use the default elasticity value only in this case.
	// since 0 is not a valid value for ALL of these parameters, we can be certain
	// the user intended to use defaults for at least those options, and won't be
	// surprised then if other defaults are used as well.
	//
	// for example, setting accelScale to a valid value (non-zero) will permit
	// configuring elasticity to 0.
	if 0 == config.AccelScale && 0 == config.Elasticity {
		f.elasticity = DefaultElasticity
	} else {
		f.elasticity = config.Elasticity
	}

	// allocate buffers for the obstacles and particles
	f.obstacle = MakeObstacles(f.width, f.height)
	f.particle = MakeParticles(f, config.NumParticles)

	// install the callback for particle movement
	f.handleMove = config.HandleMove

	return nil
}

// normalAcceleration converts and scales the raw accelerometer values to
// magnitudes usable in the logical space of Particles in the receiver Field f.
func (f *Field) normalAcceleration(x, y, z int) (int, int, int) {

	const zFactor = 8

	x *= f.accelScale
	x /= int(velocityMax)

	y *= f.accelScale
	y /= int(velocityMax)

	z *= f.accelScale
	z /= int(velocityMax) * zFactor
	if z < 0 {
		z = -z
	}

	if z >= 4 {
		z = 1
	} else {
		z = 5 - z
	}
	x -= z
	y -= z

	return x, y, z
}

// PixelIndex returns the one-dimensional index [0 .. NumPixels-1] of the given
// Position coordinates (x, y) from logical space, by first converting them to
// real Pixel coordinates in physical space.
//
// For example, after converting the given (x, y) Position coordinates to
// physical Pixel coordinates: the top-left Pixel (0, 0) would return index 0;
// The bottom-right Pixel (Width-1, Height-1) would return NumPixels-1;
// In general, the Pixel at (x, y) will return index y*Width + x.
func (f *Field) PixelIndex(x, y Position) int {
	return int(y.Dimension()*f.width + x.Dimension())
}

// Update applies a single iteration of movement to all Particles in the
// receiver Field f, based on the given raw acceleration due to gravity.
func (f *Field) Update(x, y, z int) {

	x, y, z = f.normalAcceleration(x, y, z)
	epsilon := 2*z + 1

	for i := range f.particle {
		f.particle[i].Accelerate(x, y, z, epsilon)
	}

	for i := range f.particle {
		f.particle[i].Move(f)
	}
}
