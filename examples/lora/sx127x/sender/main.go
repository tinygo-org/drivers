// Sends data with LoRa.
package main

import (
	"fmt"
	"machine"
	"strconv"
	"time"

	"tinygo.org/x/drivers/lora/sx127x"
)

var loraConfig = sx127x.Config{
	Frequency:       433998500,
	SpreadingFactor: 7,
	Bandwidth:       125000,
	CodingRate:      6,
	TxPower:         17,
}

func main() {
	println("LoRa Sender Example")

	// SPI settings for Feather M0 LoRa board
	csPin := machine.D8
	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rstPin := machine.D4
	rstPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	dio0Pin := machine.D3
	dio0Pin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// csPin := machine.P16
	// csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	// rstPin := machine.P0
	// rstPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	machine.SPI0.Configure(machine.SPIConfig{})

	loraRadio := sx127x.New(machine.SPI0, csPin, rstPin)

	var err = loraRadio.Configure(loraConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	println("Sending LoRa packets every 5 seconds...")
	for i := 0; ; i++ {
		var packet = "TinyGo LoRa Sender: " + strconv.Itoa(i)
		println("Sending:", packet)
		loraRadio.SendPacket([]byte(packet))
		time.Sleep(5000 * time.Millisecond)
	}
}
