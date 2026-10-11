// Receives data with LoRa.
package main

import (
	"fmt"
	"machine"

	"tinygo.org/x/drivers/lora/sx127x"
)

var loraConfig = sx127x.Config{
	Frequency:       433998500,
	SpreadingFactor: 7,
	Bandwidth:       125000,
	CodingRate:      6,
	TxPower:         17,
}

var packet [255]byte

func main() {
	println("LoRa Receiver Example")

	// SPI settings for Feather M0 LoRa board
	csPin := machine.D8
	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	rstPin := machine.D4
	rstPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	dio0Pin := machine.D3
	dio0Pin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	machine.SPI0.Configure(machine.SPIConfig{})
	loraRadio := sx127x.New(machine.SPI0, csPin, rstPin)
	var err = loraRadio.Configure(loraConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	println("Receiving LoRa packets...")

	for {
		packetSize := loraRadio.ParsePacket(0)
		if packetSize > 0 {
			println("Got packet, RSSI=", loraRadio.LastPacketRSSI())
			size := loraRadio.ReadPacket(packet[:])
			println(string(packet[:size]))
		}
	}
}
