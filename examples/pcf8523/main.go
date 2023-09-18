package main

import (
	"machine"
	"time"
	"tinygo.org/x/drivers/pcf8523"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})
	dev := pcf8523.New(machine.I2C0)

	// make sure the battery takes over if power is lost
	err := dev.SetPowerManagement(pcf8523.PowerManagement_SwitchOver_ModeStandard)
	if err != nil {
		panic(err)
	}

	// set RTC once, i.e. from `date -u +"%Y-%m-%dT%H:%M:%SZ"`
	now, _ := time.Parse(time.RFC3339, "2023-09-18T20:31:38Z")
	err = dev.SetTime(now)
	if err != nil {
		panic(err)
	}

	for {
		ts, err := dev.ReadTime()
		if err != nil {
			panic(err)
		}
		println("tick-tock, it's: " + ts.String())
		time.Sleep(2 * time.Second)
	}
}
