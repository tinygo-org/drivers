package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/aht20"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	dev := aht20.New(machine.I2C0)
	dev.Configure()

	dev.Reset()
	for {
		time.Sleep(500 * time.Millisecond)

		err := dev.Read()
		if err != nil {
			println("Error", err)
			continue
		}

		println("temp    ", fmtD(dev.DeciCelsius(), 3, 1), "C")
		println("humidity", fmtD(dev.DeciRelHumidity(), 3, 1), "%")
	}
}

func fmtD(val int32, i int, f int) string {
	result := make([]byte, i+f+1)
	neg := false

	if val < 0 {
		val = -val
		neg = true
	}

	for p := len(result) - 1; p >= 0; p-- {
		result[p] = byte(int32('0') + (val % 10))
		val = val / 10

		if p == i+1 && p > 0 {
			p--
			result[p] = '.'
		}
	}

	if neg {
		result[0] = '-'
	}

	return string(result)
}
