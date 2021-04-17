package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/l9110x"
)

const (
	maxSpeed = 100
)

func main() {
	machine.D11.Configure(machine.PinConfig{Mode: machine.PinOutput})
	machine.D12.Configure(machine.PinConfig{Mode: machine.PinOutput})

	err := machine.TCC0.Configure(machine.PWMConfig{})
	if err != nil {
		println(err.Error())
		return
	}

	ca, err := machine.TCC0.Channel(machine.D11)
	if err != nil {
		println(err.Error())
		return
	}

	cb, err := machine.TCC0.Channel(machine.D12)
	if err != nil {
		println(err.Error())
		return
	}

	wheel := l9110x.NewWithSpeed(ca, cb, machine.TCC0)
	wheel.Configure()

	for i := 0; i <= 10; i++ {
		println("Forward")
		var i uint32
		for i = 0; i < maxSpeed; i += 10 {
			wheel.Forward(i)
			time.Sleep(time.Millisecond * 100)
		}

		println("Stop")
		wheel.Stop()
		time.Sleep(time.Millisecond * 1000)

		println("Backward")
		for i = 0; i < maxSpeed; i += 10 {
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
