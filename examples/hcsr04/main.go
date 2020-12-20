package hcsr04

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/hcsr04"
)

func main() {
	machine.D10.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.D9.Configure(machine.PinConfig{Mode: machine.PinInput})

	sensor := hcsr04.New(machine.D10, machine.D9)
	sensor.Configure()

	println("Ultrasonic starts")
	for {
		println("Distance:", sensor.ReadDistance(), "mm")

		time.Sleep(100 * time.Millisecond)
	}
}
