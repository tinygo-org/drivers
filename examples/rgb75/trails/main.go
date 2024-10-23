package main

import (
	"machine"
	"math/rand"
	"time"

	"tinygo.org/x/drivers/rgb75"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// panel layout and color depth
	config := rgb75.Config{
		Width:      64,
		Height:     32,
		ColorDepth: 4,
		DoubleBuf:  true,
	}
	display := &screen{
		// actual rgb75 Device object
		Device: rgb75.New(
			machine.HUB75_OE, machine.HUB75_LAT, machine.HUB75_CLK,
			[6]machine.Pin{
				machine.HUB75_R1, machine.HUB75_G1, machine.HUB75_B1,
				machine.HUB75_R2, machine.HUB75_G2, machine.HUB75_B2,
			},
			[]machine.Pin{
				machine.HUB75_ADDR_A, machine.HUB75_ADDR_B, machine.HUB75_ADDR_C,
				machine.HUB75_ADDR_D, machine.HUB75_ADDR_E,
			}),
		// private data structure representing a multi-colored, continuous
		// stream of moving pixels
		trail: []*trail{
			newTrail(config.Width, 1.0),
			newTrail(config.Width, 2.0),
			newTrail(config.Width, 2.0),
			newTrail(config.Width, 0.5),
			newTrail(config.Width, 1.5),
			newTrail(config.Width, 1.5),
			newTrail(config.Width, 0.7),
			newTrail(config.Width, 1.3),
		},
	}

	if err := display.Configure(config); nil != err {
		for {
			println("error: " + err.Error())
			time.Sleep(time.Second)
		}
	}
	display.Resume()

	for {
		for _, tr := range display.trail {
			// update trail head
			if !display.contains(tr.inc()) {
				tr.wrap()
			}
			// draw each valid pixel of the trail
			for _, px := range tr.pix {
				if display.contains(px.point) {
					x, y := px.point.pos()
					display.SetPixel(x, y, px.color)
				}
			}
		}
		if err := display.Display(); nil != err {
			println("error: " + err.Error())
		}

		time.Sleep(10 * time.Millisecond)
	}
}
