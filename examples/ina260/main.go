package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/ina260"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	dev := ina260.New(machine.I2C0)
	dev.Configure(ina260.Config{
		AverageMode:     ina260.AVGMODE_16,
		VoltConvTime:    ina260.CONVTIME_140USEC,
		CurrentConvTime: ina260.CONVTIME_140USEC,
		Mode:            ina260.MODE_CONTINUOUS | ina260.MODE_VOLTAGE | ina260.MODE_CURRENT,
	})

	if dev.Connected() {
		println("INA260 detected")
	} else {
		println("INA260 NOT detected")
		return
	}

	for {
		microvolts := dev.Voltage()
		microamps := dev.Current()
		microwatts := dev.Power()

		println(fmtD(microvolts, 4, 3), "mV,", fmtD(microamps, 4, 3), "mA,", fmtD(microwatts, 4, 3), "mW")

		time.Sleep(10 * time.Millisecond)
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
