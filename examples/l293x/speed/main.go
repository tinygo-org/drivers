package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/l293x"
)

const (
	maxSpeed = 30000
)

func main() {
	machine.InitPWM()

	wheel := l293x.NewWithSpeed(machine.D10, machine.D11, machine.PWM{machine.D12})
	wheel.Configure()

	for i := 0; i <= 10; i++ {
		println("Forward")
		var i uint16
		for i = 0; i < maxSpeed; i += 1000 {
			wheel.Forward(i)
			time.Sleep(time.Millisecond * 100)
		}

		println("Stop")
		wheel.Stop()
		time.Sleep(time.Millisecond * 1000)

		println("Backward")
		for i = 0; i < maxSpeed; i += 1000 {
			wheel.Backward(i)
			time.Sleep(time.Millisecond * 100)
		}

		println("Stop")
		wheel.Stop()
		time.Sleep(time.Millisecond * 1000)
	}

	println("Stop")
	wheel.Stop()
}
