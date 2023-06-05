package main

import (
	"encoding/hex"
	"machine"
	"time"

	"tinygo.org/x/drivers/onewire"

	"tinygo.org/x/drivers/ds18b20"
)

func main() {
	// Define pin for DS18B20
	pin := machine.D2

	ow := onewire.New(pin)
	romIDs, err := ow.Search(onewire.SEARCH_ROM)
	if err != nil {
		println(err)
	}
	sensor := ds18b20.New(ow)

	for {
		time.Sleep(3 * time.Second)

		println()
		println("Device:", machine.Device)

		println()
		println("Request Temperature.")
		for _, romid := range romIDs {
			println("Sensor RomID: ", hex.EncodeToString(romid))
			sensor.RequestTemperature(romid)
		}

		// wait 750ms or more for DS18B20 convert T
		time.Sleep(1 * time.Second)

		println()
		println("Read Temperature")
		for _, romid := range romIDs {
			raw, err := sensor.ReadTemperatureRaw(romid)
			if err != nil {
				println(err)
			}
			println()
			println("Sensor RomID: ", hex.EncodeToString(romid))
			println("Temperature Raw value: ", hex.EncodeToString(raw))

			t, err := sensor.ReadTemperature(romid)
			if err != nil {
				println(err)
			}
			println("Temperature in celsius milli degrees (°C/1000): ", t)
		}
	}
}
