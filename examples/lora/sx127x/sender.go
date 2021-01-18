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
	Frequency:       868100000,
	SpreadingFactor: 12,
	Bandwidth:       125000,
	CodingRate:      6,
	TxPower:         17,
}

const (
	SPI_SCK_PIN = machine.PA5
	SPI_SDO_PIN = machine.PA7
	SPI_SDI_PIN = machine.PA6

	SPI_CS_PIN  = machine.PB8
	SPI_RST_PIN = machine.PB9
	DIO0_PIN    = machine.PA2

	UART_TX_PIN = machine.PA9
	UART_RX_PIN = machine.PA10
)

func main() {

	// Configure serial console
	machine.Serial.Configure(machine.UARTConfig{TX: UART_TX_PIN, RX: UART_RX_PIN, BaudRate: 115200})

	println("LoRa Sender Example")

	// Prepare GPIOS (SPI + DIO)
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	SPI_CS_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	SPI_RST_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	DIO0_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Enable SPI
	machine.SPI0.Configure(machine.SPIConfig{
		SCK:       SPI_SCK_PIN,
		SDO:       SPI_SDO_PIN,
		SDI:       SPI_SDI_PIN,
		Frequency: 500000,
		Mode:      0})

	// Create and Reset Lora driver
	loraRadio := sx127x.New(machine.SPI0, SPI_CS_PIN, SPI_RST_PIN)
	loraRadio.Reset()

	// Check module identification
	if loraRadio.GetVersion() != 0x12 {
		println("SX1276 module not found")
		for {
		}
	} else {
		println("Module version 0x12")
	}

	// Configure Lora settings (modulation, SF... etc )
	var err = loraRadio.SetupLora(loraConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Send a Lora packet every 5 sec
	println("Sending LoRa packets every 5 seconds...")
	for i := 0; ; i++ {
		var packet = "TinyGo LoRa Sender: " + strconv.Itoa(i)
		println("Sending:", packet)

		machine.LED.Low()
		time.Sleep(time.Millisecond * 150)
		machine.LED.High()

		loraRadio.TxLora([]byte(packet))
		time.Sleep(5000 * time.Millisecond)
	}
}
