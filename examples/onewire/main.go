package main

import (
	"encoding/hex"
	"time"
	"tinygo.org/x/drivers/onewire"
)

func main() {

	pin := machine.D2

	ow := onewire.New(pin)

	for {
		time.Sleep(3 * time.Second)

		println()
		println("Device:", machine.Device)

		romIDs, err := ow.Search(onewire.SEARCH)
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
