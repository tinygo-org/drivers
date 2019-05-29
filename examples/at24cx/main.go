package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/at24cx"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})

	eeprom := at24cx.New(machine.I2C0)
	eeprom.Configure(at24cx.Config{})

	values := make([]uint8, 100)
	for i := uint16(0); i < 100; i++ {
		values[i] = uint8(65 + i%26)
	}
	_, err := eeprom.WriteAt(values, 0)
	if err != nil {
		println("There was an error in WriteAt:", err)
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
	println("Expected: ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	print("Real: ")
	for i := uint16(0); i < 26; i++ {
		char, err := eeprom.ReadByte(i)
		print(string(char))
		if err != nil {
			println("There was an error in ReadByte:", i, err)
			return
		}
	}
	println("")

	println("\n\r\n\rRead 100 bytes from address 26")
	println("Expected: ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVWXYZABCDEFGHIJKLMNOPQRSTUVZYXWVUTSRQPONMLKJIHGFEDCBA")
	print("Real: ")
	data := make([]byte, 100)
	_, err = eeprom.ReadAt(data, 26)
	if err != nil {
		println("There was an error in ReadAt:", err)
		return
	}
	for i := 0; i < 100; i++ {
		print(string(data[i]))
	}
	println("")

	// Move to the beginning of memory
	eeprom.Seek(0, 0)
	_, err = eeprom.Write([]uint8{88, 88, 88})
	if err != nil {
		println("There was an error in Write:", err)
		return
	}

	println("\n\r\n\rRead 3 bytes")
	println("Expected: DEF")
	print("Real: ")
	data = make([]byte, 3)
	_, err = eeprom.Read(data)
	if err != nil {
		println("There was an error in Read:", err)
		return
	}
	for _, char := range data {
		print(string(char))
	}
	println("")

	println("\n\r\n\rRead another 3 bytes (from the beginning this time)")
	eeprom.Seek(-6, 1)
	println("Expected: XXX")
	print("Real: ")
	data = make([]byte, 3)
	_, err = eeprom.Read(data)
	if err != nil {
		println("There was an error in Read:", err)
		return
	}
	for _, char := range data {
		print(string(char))
	}
	println("")

	// Move to the end of memory
	eeprom.Seek(-4, 2)
	_, err = eeprom.Write([]uint8{89, 90, 89, 90})
	if err != nil {
		println("There was an error in Write:", err)
		return
	}

	println("\n\r\n\rRead the last 4 bytes of the memory and the 3 of the beginning")
	eeprom.Seek(-4, 1)
	println("Expected: YZYZXXX")
	print("Real: ")
	data = make([]byte, 7)
	_, err = eeprom.Read(data)
	if err != nil {
		println("There was an error in Read:", err)
		return
	}
	for _, char := range data {
		print(string(char))
	}
	println("")

}
