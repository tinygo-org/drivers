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
	SpreadingFactor: 7,
	Bandwidth:       125000,
	CodingRate:      6,
	TxPower:         17,
}

func main() {
	println("LoRa Sender Example")

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

	loraRadio := sx127x.New(machine.SPI0, csPin, rstPin)

	var err = loraRadio.Configure(loraConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	println("Sending LoRa packets every 5 seconds...")
	for i := 0; ; i++ {
		var packet = "TinyGo LoRa Sender: " + strconv.Itoa(counter)
		println("Sending:", packet)
		loraRadio.SendPacket([]byte(packet))
		time.Sleep(5000 * time.Millisecond)
	}
}
