package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/p1am"
)

func main() {
	for {
		if err := loop(); err != nil {
			fmt.Printf("loop failed, retrying: %v\n", err)
			time.Sleep(500 * time.Millisecond)
		}
	}
}
func loop() error {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	sw := machine.SWITCH
	sw.Configure(machine.PinConfig{Mode: machine.PinInput})

	if err := p1am.Controller.Initialize(); err != nil {
		return fmt.Errorf("initializing controller: %w", err)
	}

	version, err := p1am.Controller.Version()
	if err != nil {
		return fmt.Errorf("fetching base controller version: %w", err)
	}
	fmt.Printf("Base controller version: %d.%d.%d\n", version[0], version[1], version[2])

	for i := 1; i <= p1am.Controller.Slots; i++ {
		slot := p1am.Controller.Slot(i)
		fmt.Printf("Slot %d: ID 0x%08x, Props %+v\n", i, slot.ID, slot.Props)
	}

	slot1 := p1am.Controller.Slot(1)
	var lastInput uint32
	state := sw.Get()
	for {
		if active, err := p1am.Controller.Active(); err != nil || !active {
			return fmt.Errorf("controller active %v: %v", active, err)
		}
		if state != sw.Get() {
			state = sw.Get()
			fmt.Printf("New switch state: %v\n", state)
			if slot1.Props.DO > 0 {
				if err := slot1.Channel(1).WriteDiscrete(state); err != nil {
					return err
				}
			}
		}
		if slot1.Props.DI > 0 {
			sstate, err := slot1.ReadDiscrete()
			if err != nil {
				return fmt.Errorf("reading slot: %w", err)
			}
			if sstate != lastInput {
				lastInput = sstate
				fmt.Printf("new DI state: %#b\n", sstate)
			}
		}
		if state {
			led.High()
		} else {
			led.Low()
		}
		time.Sleep(time.Millisecond * 10)
	}
	return nil
}
