package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/onewire"
	"tinygo.org/x/drivers/onewire/devices/ds18b20"
)

func main() {
	pin := machine.D13

	ow := onewire.New(pin)
	sensor := ds18b20.New(ow)
	time.Sleep(5 * time.Second)
	for {
		println("Read OneWire ROM")
		err := sensor.ReadAddress()
		if err != nil {
			println(err)
			time.Sleep(1 * time.Second)
			continue
		}
		fmt.Printf("%x \r\n", sensor.RomID)

		println("RequestTemperature")
		err = sensor.RequestTemperature()
		if err != nil {
			println(err)
			time.Sleep(1 * time.Second)
			continue
		}

		// wait 750ms or more for DS18B20 convert T
		time.Sleep(1024 * time.Millisecond)

		println("ReadTemperature")
		temp, err := sensor.ReadTemperature()
		if err != nil {
			println(err)
			time.Sleep(1 * time.Second)
			continue
		}
		// temperature in celsius milli degrees (°C/1000)
		println("TEMP (°C/1000): ", temp)
		time.Sleep(3 * time.Second)
	}
}
