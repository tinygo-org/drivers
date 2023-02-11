package main

import (
	"time"

	"tinygo.org/x/drivers/ttp229"

	"machine"
)

func main() {
	time.Sleep(5 * time.Second)
	sensor := ttp229.NewPin(machine.A5, machine.A4)
	sensor.Configure(ttp229.Configuration{Inputs: 16})

	println("READY")
	for {
		sensor.ReadKeys()
		for i := byte(0); i < 16; i++ {
			if sensor.IsKeyPressed(i) {
				print("1 ")
			} else {
				print("0 ")
			}
		}
		println("")

		println("Pressed key:", sensor.GetKey())

		for i := byte(0); i < 16; i++ {
			if sensor.IsKeyDown(i) {
				println("Key", i, "is down")
			} else if sensor.IsKeyUp(i) {
				println("Key", i, "is up")
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}
