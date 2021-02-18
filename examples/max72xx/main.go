package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/max72xx"
)

// example for a 4 digit 7 segment display
func main() {
	// Pins for Arduino Nano 33 IOT
	err := machine.SPI0.Configure(machine.SPIConfig{
		SDO:       machine.D11, // default SDO pin
		SCK:       machine.D13, // default sck pin
		LSBFirst:  false,
		Frequency: 10000000,
	})

	if err != nil {
		println(err.Error())
	}

	driver := max72xx.NewDevice(machine.SPI0, machine.D6)
	driver.Configure()

	digitNumber := 4

	driver.StopDisplayTest()
	driver.SetDecodeMode(4)
	driver.SetScanLimit(4)
	driver.StopShutdownMode()

	for i := 1; i < int(digitNumber); i++ {
		driver.WriteCommand(byte(i), byte(Blank))
	}

	for {
		for _, character := range characters {
			println("writing", "characterValue:", character.String())
			driver.WriteCommand(byte(4), byte(character))
			driver.WriteCommand(byte(3), byte(character))
			driver.WriteCommand(byte(2), byte(character))
			driver.WriteCommand(byte(1), byte(character))
			time.Sleep(500 * time.Millisecond)

		}
		time.Sleep(time.Second)
	}

}

var characters = []Character{
	Zero,
	One,
	Two,
	Three,
	Four,
	Five,
	Six,
	Seven,
	Eight,
	Nine,
	Dash,
	E,
	H,
	L,
	P,
	Blank,
	Dot,
}

// Each bit translates to a pin, which is driven high or low
type Character byte

func (char Character) String() string {
	switch char {
	case Zero:
		return "0"
	case One:
		return "1"
	case Two:
		return "2"
	case Three:
		return "3"
	case Four:
		return "4"
	case Five:
		return "5"
	case Six:
		return "6"
	case Seven:
		return "7"
	case Eight:
		return "8"
	case Nine:
		return "9"
	case Dash:
		return "-"
	case E:
		return "E"
	case H:
		return "H"
	case L:
		return "L"
	case P:
		return "P"
	case Blank:
		return ""
	case Dot:
		return "."
	}

	return ""
}

const (
	Zero  Character = 0 //126
	One   Character = 1 //48
	Two   Character = 2 // 109
	Three Character = 3 // 121
	Four  Character = 4
	Five  Character = 5
	Six   Character = 6
	Seven Character = 7
	Eight Character = 8
	Nine  Character = 9
	Dash  Character = 10
	E     Character = 11
	H     Character = 12
	L     Character = 13
	P     Character = 14
	Blank Character = 15
	Dot   Character = 128
)
