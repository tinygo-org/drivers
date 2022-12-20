//go:build rp2040

package ssd1289

import (
	"device/rp"
	"machine"
)

type rp2040Bus struct {
	firstPin machine.Pin
}

func NewRP2040Bus(firstPin machine.Pin) rp2040Bus {

	for i := uint8(0); i < 16; i++ {
		pin := machine.Pin(i + uint8(firstPin))
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}

	return rp2040Bus{
		firstPin: firstPin,
	}

}

func (b rp2040Bus) Set(data uint16) {
	data32 := uint32(data)
	rp.SIO.GPIO_OUT_CLR.Set(0xFFFF << b.firstPin)
	rp.SIO.GPIO_OUT_SET.Set(data32 << b.firstPin)
}
