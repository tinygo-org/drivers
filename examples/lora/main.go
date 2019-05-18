// Sends data with LoRa.
package main

import (
	"fmt"
	"machine"
	"strconv"
	"time"

	"github.com/tinygo-org/drivers/lora"
)

func main() {
	println("LoRa Example")
	time.Sleep(10000 * time.Millisecond)
	println("LoRa Example 1")
	time.Sleep(10000 * time.Millisecond)
	// SPI settings for Feather M0 LoRa board
	csPin := machine.GPIO{machine.D8}
	csPin.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	rstPin := machine.GPIO{machine.D4}
	rstPin.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	dio0Pin := machine.GPIO{machine.D3}
	dio0Pin.Configure(machine.GPIOConfig{Mode: machine.GPIO_INPUT})

	println("LoRa Example 2")
	machine.SPI0.Configure(machine.SPIConfig{})

	println("LoRa Example 3")
	transceiver := lora.New(machine.SPI0, csPin, rstPin, dio0Pin)
	println("LoRa Example 4")
	var err = transceiver.Configure(lora.Config{})
	transceiver.PrintRegisters()
	if err != nil {
		fmt.Println(err)
		return
	}

	var counter int = 0
	for {
		println("LoRa Example Loop")
		counter += 1
		var packet = "TinyGo LoRA: " + strconv.Itoa(counter)
		transceiver.SendPacket([]byte(packet))
		time.Sleep(2000 * time.Millisecond)
	}
}
