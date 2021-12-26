package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/shtc3"
)

func main() {

	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := shtc3.New(machine.I2C0)

	for {

		sensor.WakeUp()

		temp, humidity, _ := sensor.ReadTemperatureHumidity()
		t := fmt.Sprintf("%.2f", float32(temp)/1000)
		h := fmt.Sprintf("%.2f", float32(humidity)/100)
		println("Temperature:", t, "°C")
		println("Humidity", h, "%")

		sensor.Sleep()

		time.Sleep(2 * time.Second)
	}
}
