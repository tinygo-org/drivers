package main

import (
	"machine"
	"time"

	"github.com/tinygo-org/drivers/at24cx"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	eeprom := at24cx.New(machine.I2C0)
	eeprom.Configure(at24cx.Config{})

	values := make([]uint8, 100)
	for i := uint16(0); i < 100; i++ {
		values[i] = uint8(65 + i%26)
	}
	err := eeprom.WriteBytes(0, values)
	if err != nil {
		println("There was an error in WriteBytes:", err)
		return
	}

	for i := uint16(0); i < 26; i++ {
		err = eeprom.WriteByte(100+i, uint8(90-i))
		if err != nil {
			println("There was an error in WriteByte:", i, err)
			return
		}
		time.Sleep(2 * time.Millisecond)
	}

	println("\n\r\n\rRead 26 bytes one by one from address 0")
	for i := uint16(0); i < 26; i++ {
		char, err := eeprom.ReadByte(i)
		print(string(char))
		if err != nil {
			println("There was an error in ReadByte:", i, err)
			return
		}
	}

	println("\n\r\n\rRead 100 bytes from address 26")
	data := make([]byte, 100)
	err = eeprom.ReadBytes(26, data)
	if err != nil {
		println("There was an error in ReadBytes:", err)
		return
	}
	for i := 0; i < 100; i++ {
		print(string(data[i]))
	}
}
