package main

import (
	"encoding/hex"
	"machine"
	"time"

	onewire "tinygo.org/x/drivers/onewire_v2"
)

type onewirePin struct {
	p        machine.Pin
	isOutput bool
}

func (owp *onewirePin) Set(level bool) {
	if level && owp.isOutput {
		owp.p.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
		owp.isOutput = false
	}
	if !level && !owp.isOutput {
		owp.p.Configure(machine.PinConfig{Mode: machine.PinOutput})
		owp.isOutput = true
	}
}

func (owp *onewirePin) Get() bool {
	if owp.isOutput {
		owp.p.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
		owp.isOutput = false
	}
	return owp.p.Get()
}

func main() {
	pin := &onewirePin{p: machine.D2, isOutput: false}
	ow := onewire.New(pin)

	for {
		time.Sleep(3 * time.Second)

		println()
		println("Device:", machine.Device)

		romIDs, err := ow.Search(onewire.SEARCH_ROM)
		if err != nil {
			println(err)
		}
		for _, romid := range romIDs {
			println(hex.EncodeToString(romid))
		}

		if len(romIDs) == 1 {
			// only 1 device on bus
			r, err := ow.ReadAddress()
			if err != nil {
				println(err)
			}
			println(hex.EncodeToString(r))

		}
	}
}
