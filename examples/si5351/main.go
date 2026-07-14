package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/si5351"
)

// Simple demo of the SI5351 clock generator.
// This is like the Arduino library example:
//   https://github.com/adafruit/Adafruit_Si5351_Library/blob/master/examples/si5351/si5351.ino
// Which will configure the chip with:
//  - PLL A at 900mhz
//  - PLL B at 616.66667mhz
//  - Clock 0 at 112.5mhz, using PLL A as a source divided by 8
//  - Clock 1 at 13.5531mhz, using PLL B as a source divided by 45.5
//  - Clock 2 at 10.76khz, using PLL B as a source divided by 900 and further divided with an R divider of 64.

func main() {
	time.Sleep(5 * time.Second)

	println("Si5351 Clockgen Test")
	println()

	// Configure I2C bus
	machine.I2C0.Configure(machine.I2CConfig{})

	// Create driver instance
	clockgen := si5351.New(machine.I2C0)

	// Initialize device
	cnf := si5351.Config{
		Capacitance: si5351.CrystalLoad10PF,
	}

	if err := clockgen.Configure(cnf); err != nil {
		println("Failed to configure Si5351:", err.Error())
		return
	}
	println("Si5351 configured")

	// Now configure the clock outputs.
	clockgen.SetFrequency(si5351.Clock0, 112_500_000)
	println("Clock 0: 112.5mhz")

	// Next configure clock 1 for 13.5531mhz (616.6667mhz / 45.5).
	// This uses fractional division.
	clockgen.SetFrequency(si5351.Clock1, 13_553_125)
	println("Clock 1: 13.5531mhz")

	// Finally configure clock 2 to output of 10.706khz.
	clockgen.SetFrequency(si5351.Clock2, 10_706)
	println("Clock 2: 10.706khz")

	// After configuring the clocks enable the outputs.
	clockgen.EnableOutput(si5351.Clock0, true)
	clockgen.EnableOutput(si5351.Clock1, true)
	clockgen.EnableOutput(si5351.Clock2, true)
	println("All outputs enabled")

	time.Sleep(time.Second)

	clockgen.EnableOutput(si5351.Clock0, false)
	clockgen.EnableOutput(si5351.Clock1, false)
	clockgen.EnableOutput(si5351.Clock2, false)
	println("All outputs disabled for 5 seconds")
	time.Sleep(5 * time.Second)

	// Now turn clock outputs on and off repeatedly
	on := false
	for {
		if on {
			println("Setting clock outputs off")
			clockgen.EnableOutput(si5351.Clock0, false)
			clockgen.EnableOutput(si5351.Clock1, false)
			clockgen.EnableOutput(si5351.Clock2, false)
			on = false
		} else {
			println("Setting clock outputs on")
			clockgen.EnableOutput(si5351.Clock0, true)
			clockgen.EnableOutput(si5351.Clock1, true)
			clockgen.EnableOutput(si5351.Clock2, true)
			on = true
		}
		time.Sleep(1 * time.Second)
	}
}
