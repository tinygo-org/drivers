// Draws the Lorenz attractor on the ITead ITDB02-2.8 shield. A 3000 point
// trail traces the attractor while the view slowly rotates around the z axis.
package main

import (
	"image/color"
	"machine"
	"math"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/ili9325"
)

var (
	data = [8]machine.Pin{machine.D0, machine.D1, machine.D2, machine.D3, machine.D4, machine.D5, machine.D6, machine.D7}
	rs   = machine.ADC5
	wr   = machine.ADC4
	cs   = machine.ADC3
	rst  = machine.ADC2
)

const (
	trail = 3000
	sigma = 10
	rho   = 28
	beta  = 8.0 / 3
	dt    = 0.006
	scale = 5.2
)

var (
	black   = color.RGBA{0, 0, 0, 255}
	palette [64]color.RGBA
	px, py  [trail]int16
)

func main() {
	out := machine.PinConfig{Mode: machine.PinOutput}
	for _, p := range append(data[:], rs, wr, cs, rst) {
		p.Configure(out)
	}

	display := ili9325.New(newBus(), rs, cs, rst)
	display.Configure(ili9325.Config{Rotation: drivers.Rotation90})
	display.FillScreen(black)

	for i := range palette {
		palette[i] = hue(float32(i) / float32(len(palette)))
	}
	for i := range px {
		px[i] = -1
	}

	w, h := display.Size()
	cx, cy := float32(w/2), float32(h-6)
	x, y, z := float32(0.1), float32(0), float32(0)
	theta := float64(0)
	cos, sin := float32(1), float32(0)
	for i := 0; ; i++ {
		dx := sigma * (y - x)
		dy := x*(rho-z) - y
		dz := x*y - beta*z
		x += dx * dt
		y += dy * dt
		z += dz * dt

		if i%64 == 0 {
			theta += 0.004
			cos, sin = float32(math.Cos(theta)), float32(math.Sin(theta))
		}
		sx := int16(cx + (x*cos-y*sin)*scale)
		sy := int16(cy - z*4.4)

		j := i % trail
		if px[j] >= 0 {
			display.SetPixel(px[j], py[j], black)
		}
		zi := int(z * 1.3)
		if zi < 0 {
			zi = 0
		}
		display.SetPixel(sx, sy, palette[zi%len(palette)])
		px[j], py[j] = sx, sy
	}
}

// hue returns a fully saturated colour for h in [0, 1).
func hue(h float32) color.RGBA {
	h6 := h * 6
	f := h6 - float32(int(h6))
	up, down := uint8(255*f), uint8(255*(1-f))
	switch int(h6) % 6 {
	case 0:
		return color.RGBA{255, up, 0, 255}
	case 1:
		return color.RGBA{down, 255, 0, 255}
	case 2:
		return color.RGBA{0, 255, up, 255}
	case 3:
		return color.RGBA{0, down, 255, 255}
	case 4:
		return color.RGBA{up, 0, 255, 255}
	default:
		return color.RGBA{255, 0, down, 255}
	}
}
