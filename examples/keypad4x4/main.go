package main

import (
	"machine"

	"tinygo.org/x/drivers/keypad4x4"
)

func main() {
	mapping := map[uint8]string{
		1:  "1",
		2:  "2",
		3:  "3",
		4:  "A",
		5:  "4",
		6:  "5",
		7:  "6",
		8:  "B",
		9:  "7",
		10: "8",
		11: "9",
		12: "C",
		13: "*",
		14: "0",
		15: "#",
		16: "D",
	}

	keypadDevice := keypad4x4.NewDevice(machine.D2, machine.D3, machine.D4, machine.D5, machine.D6, machine.D7, machine.D8, machine.D9)
	keypadDevice.Configure()

	for {
		key := keypadDevice.GetKey()
		if key != keypad4x4.NoKeyPressed {
			println("Button: ", mapping[key])
		}
	}
}
