package main

import (
	"machine"
	"strconv"
	"time"

	"tinygo.org/x/drivers/hd44780i2c"
)

func main() {

	// Note: most HD44780 LCD modules requires 5V power, however some variations
	// use 3.3V (and may be damaged by 5V).

	machine.I2C0.Configure(machine.I2CConfig{
		Frequency: machine.TWI_FREQ_400KHZ,
	})

	lcd := hd44780i2c.New(machine.I2C0, 0x27) // some modules have address 0x3F

	lcd.Configure(hd44780i2c.Config{
		Width:       16, // required
		Height:      2,  // required
		CursorOn:    true,
		CursorBlink: true,
	})

	lcd.Print([]byte(" TinyGo\n  LCD Test "))

	// CGRAM address 0x0-0x7 can be used to store 8 custom characters
	lcd.CreateCharacter(0x0, []byte{0x00, 0x11, 0x0E, 0x1F, 0x15, 0x1F, 0x1F, 0x1F})
	lcd.Print([]byte{0x0})

	// You can use https://maxpromer.github.io/LCD-Character-Creator/
	// to crete your own characters.

	time.Sleep(time.Millisecond * 7000)

	for i := 0; i < 5; i++ {
		lcd.BacklightOn(false)
		time.Sleep(time.Millisecond * 250)
		lcd.BacklightOn(true)
		time.Sleep(time.Millisecond * 250)
	}

	lcd.CursorOn(false)
	lcd.CursorBlink(false)

	i := 0
	for {

		lcd.ClearDisplay()
		lcd.SetCursor(2, 1)
		lcd.Print([]byte(strconv.FormatInt(int64(i), 10)))
		i++
		time.Sleep(time.Millisecond * 100)

	}
}
