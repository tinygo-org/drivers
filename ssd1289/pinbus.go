package ssd1289

import "machine"

type pinBus struct {
	pins [16]machine.Pin
}

func NewPinBus(pins [16]machine.Pin) pinBus {

	for i := 0; i < 16; i++ {
		pins[i].Configure(machine.PinConfig{Mode: machine.PinOutput})
	}

	return pinBus{
		pins: pins,
	}
}

func (b pinBus) Set(data uint16) {
	b.pins[15].Set((data & (1 << 15)) != 0)
	b.pins[14].Set((data & (1 << 14)) != 0)
	b.pins[13].Set((data & (1 << 13)) != 0)
	b.pins[12].Set((data & (1 << 12)) != 0)
	b.pins[11].Set((data & (1 << 11)) != 0)
	b.pins[10].Set((data & (1 << 10)) != 0)
	b.pins[9].Set((data & (1 << 9)) != 0)
	b.pins[8].Set((data & (1 << 8)) != 0)
	b.pins[7].Set((data & (1 << 7)) != 0)
	b.pins[6].Set((data & (1 << 6)) != 0)
	b.pins[5].Set((data & (1 << 5)) != 0)
	b.pins[4].Set((data & (1 << 4)) != 0)
	b.pins[3].Set((data & (1 << 3)) != 0)
	b.pins[2].Set((data & (1 << 2)) != 0)
	b.pins[1].Set((data & (1 << 1)) != 0)
	b.pins[0].Set((data & (1 << 0)) != 0)
}
