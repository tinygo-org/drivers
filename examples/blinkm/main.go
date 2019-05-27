// Connects to an BlinkM I2C RGB LED.
// http://thingm.com/fileadmin/thingm/downloads/BlinkM_datasheet.pdf
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/blinkm"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	bm := blinkm.New(machine.I2C0)
	bm.Configure()

	maj, min, _ := bm.Version()

	println("Firmware version:", string(maj), string(min))

	count := 0
	for {
		switch count {
		case 0:
			// Crimson
			bm.SetRGB(0xdc, 0x14, 0x3c)
			count = 1
		case 1:
			// MediumPurple
			bm.SetRGB(0x93, 0x70, 0xdb)
			count = 2
		case 2:
			// MediumSeaGreen
			bm.SetRGB(0x3c, 0xb3, 0x71)
			count = 0
		}

		time.Sleep(100 * time.Millisecond)
	}
}
