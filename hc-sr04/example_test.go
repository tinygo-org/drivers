package main

import (
	"machine"
	"time"
)

func ExampleBasic() {
	const (
		echo = machine.GP3
		out  = machine.GP2
	)
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	d := New(out, echo, 100*time.Millisecond)
	t := time.Now()
	for {
		time.Sleep(time.Second)
		println(time.Since(t).Milliseconds())
		t = time.Now()
		led.Set(!led.Get())
		dist := d.ReadDistance()
		print(dist)
		println("mm")
	}
}
