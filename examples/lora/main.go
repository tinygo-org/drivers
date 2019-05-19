// Sends data with LoRa.
package main

import (
	"fmt"
	"machine"
	"strconv"
	"time"

	"github.com/tinygo-org/drivers/lora"
)

var loraConfig = lora.Config{
	Frequency:       433998500,
	SpreadingFactor: 9,
	Bandwidth:       312500,
	CodingRate:      6,
	TxPower:         17,
}

func main() {
	println("LoRa Example")

	// SPI settings for Feather M0 LoRa board
	// csPin := machine.GPIO{machine.D8}
	// csPin.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	// rstPin := machine.GPIO{machine.D4}
	// rstPin.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	// dio0Pin := machine.GPIO{machine.D3}
	// dio0Pin.Configure(machine.GPIOConfig{Mode: machine.GPIO_INPUT})

	csPin := machine.GPIO{machine.P16}
	csPin.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	rstPin := machine.GPIO{machine.P0}
	rstPin.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})

	machine.SPI0.Configure(machine.SPIConfig{})

	transceiver := lora.New(machine.SPI0, csPin, rstPin)

	var err = transceiver.Configure(loraConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	var counter int = 0
	for {
		counter += 1
		var packet = "TinyGo LoRa: " + strconv.Itoa(counter)
		println("Sending:", packet)
		transceiver.SendPacket([]byte(packet))
		time.Sleep(5000 * time.Millisecond)
	}
}
