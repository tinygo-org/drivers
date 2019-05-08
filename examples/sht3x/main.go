package main

import (
	"machine"
	"time"

	"github.com/tinygo-org/drivers/sht3x"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := sht3x.New(machine.I2C0)

	for {
		// sensor.Read()
		temp, humidity := sensor.Read()
		// t := fmt.Sprintf("%.2f", uint16(temp))
		// h := fmt.Sprintf("%.2f", uint16(humidity))
		//
		println(uint16(temp))
		println(uint16(humidity))
		// println(t > h)
		// println("Temperature:", t, "ºC")
		// println("Humidity", h, "%")
		time.Sleep(2 * time.Second)
	}
}
