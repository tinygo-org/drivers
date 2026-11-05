//go:build atsamd21

package main

import "tinygo.org/x/drivers/ili9325"

func newBus() ili9325.Bus {
	bus, err := ili9325.NewPortBus(data, wr)
	if err != nil {
		panic(err)
	}
	return bus
}
