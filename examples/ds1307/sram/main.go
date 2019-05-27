package main

import (
	"machine"

	"tinygo.org/x/drivers/ds1307"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})
	rtc := ds1307.New(machine.I2C0)
	read := make([]byte, 5)
	for {
		rtc.Seek(0, 0)
		_, err := rtc.Write([]byte{1, 2, 3, 4, 5})
		if err != nil {
			println("Error while writing data:", err)
			break
		}
		rtc.Seek(0, 0)
		_, err = rtc.Read(read)
		if err != nil {
			println("Error while reading data:", err)
			break
		}
		for data := range read {
			println(data, " ")
		}

	}

}
