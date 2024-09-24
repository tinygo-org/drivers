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

	// Verify device wired properly
	if !clockgen.Connected() {
		for {
			println("Ooops, no Si5351 detected ... Check your wiring!")
			time.Sleep(time.Second)
		}
	}

	// Initialise device
	clockgen.Configure()

	// Now configue the PLLs and clock outputs.
	// The PLLs can be configured with a multiplier and division of the on-board
	// 25mhz reference crystal.  For example configure PLL A to 900mhz by multiplying
	// by 36.  This uses an integer multiplier which is more accurate over time
	// but allows less of a range of frequencies compared to a fractional
	// multiplier shown next.
	clockgen.ConfigurePLL(si5351.PLL_A, 36, 0, 1) // Multiply 25mhz by 36
	println("PLL A frequency: 900mhz")

	// And next configure PLL B to 616.6667mhz by multiplying 25mhz by 24.667 using
	// the fractional multiplier configuration.  Notice you specify the integer
	// multiplier and then a numerator and denominator as separate values, i.e.
	// numerator 2 and denominator 3 means 2/3 or 0.667.  This fractional
	// configuration is susceptible to some jitter over time but can set a larger
	// range of frequencies.
	clockgen.ConfigurePLL(si5351.PLL_B, 24, 2, 3) // Multiply 25mhz by 24.667 (24 2/3)
	println("PLL B frequency: 616.6667mhz")

	// Now configure the clock outputs.  Each is driven by a PLL frequency as input
	// and then further divides that down to a specific frequency.
	// Configure clock 0 output to be driven by PLL A divided by 8, so an output
	// of 112.5mhz (900mhz / 8).  Again this uses the most precise integer division
	// but can't set as wide a range of values.
	clockgen.ConfigureMultisynth(0, si5351.PLL_A, 8, 0, 1) // Divide by 8 (8 0/1)
	println("Clock 0: 112.5mhz")

	// Next configure clock 1 to be driven by PLL B divided by 45.5 to get
	// 13.5531mhz (616.6667mhz / 45.5).  This uses fractional division and again
	// notice the numerator and denominator are explicitly specified.  This is less
	// precise but allows a large range of frequencies.
	clockgen.ConfigureMultisynth(1, si5351.PLL_B, 45, 1, 2) // Divide by 45.5 (45 1/2)
	println("Clock 1: 13.5531mhz")

	// Finally configure clock 2 to be driven by PLL B divided once by 900 to get
	// down to 685.15 khz and then further divided by a special R divider that
	// divides 685.15 khz by 64 to get a final output of 10.706khz.
	clockgen.ConfigureMultisynth(2, si5351.PLL_B, 900, 0, 1) // Divide by 900 (900 0/1)
	// Set the R divider, this can be a value of:
	//  - R_DIV_1: divider of 1
	//  - R_DIV_2: divider of 2
	//  - R_DIV_4: divider of 4
	//  - R_DIV_8: divider of 8
	//  - R_DIV_16: divider of 16
	//  - R_DIV_32: divider of 32
	//  - R_DIV_64: divider of 64
	//  - R_DIV_128: divider of 128
	clockgen.ConfigureRdiv(2, si5351.R_DIV_64)
	println("Clock 2: 10.706khz")

	// After configuring PLLs and clocks, enable the outputs.
	clockgen.EnableOutputs()

	for {
		time.Sleep(5 * time.Second)
		println()
		println("Clock 0: 112.5mhz")
		println("Clock 1: 13.5531mhz")
		println("Clock 2: 10.706khz")
	}

}
