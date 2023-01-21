package main

import (
	"machine"
	"time"
	wire "tinygo.org/x/drivers/1-wire"

	"tinygo.org/x/drivers/ds18b20"
)

func main() {
	pin := machine.D2

	ow := wire.New(pin)
	sensor := ds18b20.New(ow)
	for {
		time.Sleep(3 * time.Second)

		println()
		println("Device:", machine.Device)

		println("Read 1-Wire ROM.")
		println("Send command =", SliceToHexString([]uint8{wire.ONEWIRE_READ_ROM}))
		err := sensor.ReadAddress()
		if err != nil {
			println(err)
			continue
		}
		println("RomID:", SliceToHexString(sensor.RomID))

		println("Request Temperature.")
		println("Send command =", SliceToHexString([]uint8{ds18b20.DS18B20_CONVERT_TEMPERATURE}))
		err = sensor.RequestTemperature()
		if err != nil {
			println(err)
			continue
		}

		// wait 750ms or more for DS18B20 convert T
		time.Sleep(1 * time.Second)

		println("Read Temperature")
		println("Send command =", SliceToHexString([]uint8{ds18b20.DS18B20_READ_SCRATCHPAD}))
		t, err := sensor.ReadTemperature()
		if err != nil {
			println(err)
			continue
		}
		// temperature in celsius milli degrees (°C/1000)
		println("Temperature (°C/1000): ", t)
	}
}

// SliceToHexString converts a slice to Hex string
// fmt.Printf - compile error on an Arduino Uno boards
func SliceToHexString(rom []uint8) string {
	const hc string = "0123456789ABCDEF"
	var result string = "0x"
	for _, v := range rom {
		if v < 0x10 {
			result += "0" + string(hc[v])
		} else {
			result += string(hc[v&0xF0>>4]) + string(hc[v&0x0F])
		}
	}
	return result
}
