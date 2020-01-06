package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/l9110x"
)

func main() {
	wheel := l9110x.New(machine.D10, machine.D11)
	wheel.Configure()

	for i := 0; i <= 10; i++ {
		println("Forward")
		wheel.Forward()
		time.Sleep(time.Millisecond * 1000)

		println("Stop")
		wheel.Stop()
		time.Sleep(time.Millisecond * 1000)

		println("Backward")
		wheel.Backward()
		time.Sleep(time.Millisecond * 1000)

		println("Stop")
		wheel.Stop()
		time.Sleep(time.Millisecond * 1000)
	}

	println("Stop")
	wheel.Stop()
}
