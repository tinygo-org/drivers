//go:build !atsamd21

package main

import "tinygo.org/x/drivers/ili9325"

func newBus() ili9325.Bus {
	return ili9325.NewPinBus(data[0], data[1], data[2], data[3], data[4], data[5], data[6], data[7], wr)
}
