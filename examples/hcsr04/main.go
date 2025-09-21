package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/hcsr04"
	"tinygo.org/x/drivers/tinygo"
)

func main() {
	trigger := tinygo.New(machine.D10) // automatically configures pin as output
	echo := tinygo.New(machine.D9)     // automatically configures pin as input
	sensor := hcsr04.New(trigger, echo)
	sensor.Configure()

	println("Ultrasonic starts")
	for {
		println("Distance:", sensor.ReadDistance(), "mm")

		time.Sleep(100 * time.Millisecond)
	}
}
