package main

import (
	"image/color"
	"machine"
	"math/rand"
	"time"

	"tinygo.org/x/drivers/examples/rgb75/gravity/accel"
	"tinygo.org/x/drivers/examples/rgb75/gravity/image"

	"tinygo.org/x/drivers/lis3dh"
	"tinygo.org/x/drivers/rgb75"
)

// initAccelerometer initializes required peripherals, configures the I2C
// interface and sensor, and returns a new Accel object.
func initAccelerometer() (*accel.Accel, error) {

	// I2C interface of accelerometer
	machine.I2C0.Configure(machine.I2CConfig{
		SCL: machine.I2C0_SCL_PIN,
		SDA: machine.I2C0_SDA_PIN,
	})

	// accelerometer address and sensitivity
	config := accel.Config{
		Address: lis3dh.Address1,
		Range:   lis3dh.RANGE_4_G,
	}

	// accelerometer object
	acc := accel.New(machine.I2C0)

	if err := acc.Configure(config); nil != err {
		return nil, err
	}

	return acc, nil
}

// initDisplay initializes required peripherals, configures the HUB75 interface,
// and returns a new rgb75 device object.
func initDisplay() (*rgb75.Device, error) {

	// panel layout and color depth
	config := rgb75.Config{
		Width:      64,
		Height:     32,
		ColorDepth: 4,
		DoubleBuf:  true,
	}

	// rgb75 Device object
	hub := rgb75.New(
		machine.HUB75_OE, machine.HUB75_LAT, machine.HUB75_CLK,
		[6]machine.Pin{
			machine.HUB75_R1, machine.HUB75_G1, machine.HUB75_B1,
			machine.HUB75_R2, machine.HUB75_G2, machine.HUB75_B2,
		},
		[]machine.Pin{
			machine.HUB75_ADDR_A, machine.HUB75_ADDR_B, machine.HUB75_ADDR_C,
			machine.HUB75_ADDR_D, machine.HUB75_ADDR_E,
		})

	if err := hub.Configure(config); nil != err {
		return nil, err
	}

	return hub, nil
}

// initField initializes the buffers and logical spaces used for command and
// control of the particles.
func initField(hub *rgb75.Device) (*Field, error) {

	// Field configuration
	config := FieldConfig{
		Width:        64,
		Height:       32,
		NumParticles: 64,
		AccelScale:   1,
		Elasticity:   128,
		HandleMove: func(f *Field, p *Particle, x, y Position) {
			// clear previous pixel
			hub.SetPixel(
				int16(p.x.Dimension()),
				int16(p.y.Dimension()),
				rgb75.ClearColor,
			)
			var c color.RGBA
			if p.index&1 == 0 {
				c = color.RGBA{R: 0x00, G: 0xFF, B: 0x00, A: 0xFF}
			} else {
				c = color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF}
			}
			// set new pixel
			hub.SetPixel(
				int16(x.Dimension()),
				int16(y.Dimension()),
				c,
			)
		},
	}

	// Field object
	field := NewField()

	if err := field.Configure(config); nil != err {
		return nil, err
	}

	return field, nil
}

// drawImage draws the background image that acts as an obstacle to particles.
func drawImage(hub *rgb75.Device, field *Field) {
	w, h := image.Gopher.Size()
	for y := int16(0); y < h; y++ {
		for x := int16(0); x < w; x++ {
			if c := image.Gopher.ColorAt(x, y); 0 != c.A {
				// the image and display have opposite (x, y) axes
				hub.SetPixel(y, x, c)
				field.obstacle.Set(Dimension(y), Dimension(x))
			}
		}
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	acc, err := initAccelerometer()
	if nil != err {
		halt(err)
	}

	hub, err := initDisplay()
	if nil != err {
		halt(err)
	}

	field, err := initField(hub)
	if nil != err {
		halt(err)
	}

	hub.Resume()

	for {
		//println(acc.String())
		drawImage(hub, field)

		if x, y, z, err := acc.Get(); nil == err {
			field.Update(x, y, z)
		}

		if err := hub.Display(); nil != err {
			halt(err)
		}

		//time.Sleep(10 * time.Millisecond)
	}
}

// halt terminates program execution by continuously printing the given error
// periodically and never returning.
func halt(err error) {
	for {
		println("error: " + err.Error())
		time.Sleep(time.Second)
	}
}
