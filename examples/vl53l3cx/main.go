// This example uses the bottom VL53L3CX of an M5Stack StampFly. Both sensors
// on that board start at the same I2C address, so the front one is held in reset.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/vl53l3cx"
)

const (
	xshutBottom = machine.GPIO7
	xshutFront  = machine.GPIO9
)

func main() {
	time.Sleep(2 * time.Second)

	xshutBottom.Configure(machine.PinConfig{Mode: machine.PinOutput})
	xshutFront.Configure(machine.PinConfig{Mode: machine.PinOutput})
	xshutFront.Low()
	xshutBottom.High()
	time.Sleep(10 * time.Millisecond)

	machine.I2C0.Configure(machine.I2CConfig{
		SDA:       machine.GPIO3,
		SCL:       machine.GPIO4,
		Frequency: 400 * machine.KHz,
	})

	sensor := vl53l3cx.New(machine.I2C0)
	if !sensor.Connected() {
		println("VL53L3CX not connected")
		return
	}

	if err := sensor.Configure(vl53l3cx.Config{Mode: vl53l3cx.Medium}); err != nil {
		println("VL53L3CX configuration failed:", err.Error())
		return
	}

	if err := sensor.StartRanging(); err != nil {
		println("VL53L3CX start failed:", err.Error())
		return
	}

	var m vl53l3cx.Measurement
	for {
		if err := sensor.Read(&m); err != nil {
			println("Error reading:", err.Error())
			continue
		}
		print("targets ", m.NumTargets)
		for i := 0; i < int(m.NumTargets); i++ {
			t := &m.Targets[i]
			print(" | ", t.Distance, " mm status ", t.Status, " signal ", t.SignalRate, " kcps")
		}
		println()
	}
}
