package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/sht4x"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := sht4x.New(machine.I2C0)

	for {
		temp, humidity, _ := sensor.ReadTemperatureHumidity()
		t := fmt.Sprintf("%.2f", float32(temp)/1000)
		h := fmt.Sprintf("%.2f", float32(humidity)/100)
		println("Temperature: ", t, "°C")
		println("Humidity:    ", h, "%")
		time.Sleep(2 * time.Second)
	}
}
