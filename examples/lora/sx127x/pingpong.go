// Receives data with LoRa.
package main

import (
	"fmt"
	"machine"
	"strconv"
	"strings"

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

	// Serial console
	UART_TX_PIN = machine.PA9
	UART_RX_PIN = machine.PA10

	// RFM95 SPI Connection to Bluepill
	SPI_SCK_PIN = machine.PA5
	SPI_SDO_PIN = machine.PA7
	SPI_SDI_PIN = machine.PA6
	SPI_CS_PIN  = machine.PB8
	SPI_RST_PIN = machine.PB9

	// DIO RFM95 Pin connection to BluePill
	DIO0_PIN        = machine.PA2
	DIO0_PIN_MODE   = machine.PinInputPulldown
	DIO0_PIN_CHANGE = machine.PinRising

	pollDelayMs int = 1000
)

// configureDioInt sets up the DIO0 interrupt handler
func configureDioInt(radio *sx127x.Device) {
	// Set an interrupt on this pin.
	err := DIO0_PIN.SetInterrupt(DIO0_PIN_CHANGE, func(machine.Pin) {
		if DIO0_PIN.Get() {
			radio.CheckIrq()
		}
	})
	if err != nil {
		println("could not configure pin interrupt:", err.Error())
	}
}

func main() {
	// Configure serial console
	machine.Serial.Configure(machine.UARTConfig{TX: UART_TX_PIN, RX: UART_RX_PIN, BaudRate: 115200})

	println("LoRa PingPong Example")

	// Prepare GPIOS (SPI + DIO)
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	SPI_CS_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	SPI_RST_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	DIO0_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})

	machine.LED.Set(true)

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

	// Setup DIO0 interrupt Handling
	configureDioInt(&loraRadio)

	// Configure Lora settings (modulation, SF... etc )
	var err = loraRadio.SetupLora(loraConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	// ping counter is incremented after every retransmission
	pingCount := 0

	// Send first packet without using interrupts
	println("*** Send a Lora PING ")
	machine.LED.Set(false)
	loraRadio.TxLora([]byte("PING " + strconv.Itoa(pingCount)))
	machine.LED.Set(true)

	println("*** Switch to RX Mode ")
	for {
		loraRadio.RxLora()
		// Wait for Radio event
		in := <-loraRadio.GetRadioEventChan()

		// We have a packet
		if in.EventType == sx127x.EventRxDone {
			data := string(in.EventData)
			println("RX: ", data)

			machine.LED.Set(false)
			// If it's ping, reply pong
			if strings.Contains(data, "PING") {
				println("Received PING, Sending PONG")
				loraRadio.TxLora([]byte("PONG #" + strconv.Itoa(pingCount)))
			} else if strings.Contains(data, "PONG") {
				println("Received PONG, Sending PING")
				loraRadio.TxLora([]byte("PING #" + strconv.Itoa(pingCount)))
			}
			machine.LED.Set(true)

			pingCount++

		}

	}
}
