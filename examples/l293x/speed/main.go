package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/l293x"
)

const (
	maxSpeed = 100
)

func main() {
	err := machine.TCC0.Configure(machine.PWMConfig{
		Period: 16384e3, // 16.384ms
	})
	if err != nil {
		println(err.Error())
		return
	}

	spc, err := machine.TCC0.Channel(machine.D12)
	if err != nil {
		println(err.Error())
		return
	}

	wheel := l293x.NewWithSpeed(machine.D10, machine.D11, spc, machine.TCC0)
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
