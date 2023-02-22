package main

// This example code demonstrates Lora RX/TX With SX127x driver
// You need to connect SPI, RST, CS, DIO0 (aka IRQ) and DIO1 to use.

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lora"
	"tinygo.org/x/drivers/sx127x"
)

const (
	LORA_DEFAULT_RXTIMEOUT_MS = 1000
	LORA_DEFAULT_TXTIMEOUT_MS = 5000
)

var (
	loraRadio *sx127x.Device
	txmsg     = []byte("Hello TinyGO")

	// We assume LoRa Featherwing module is connected to PyBadge:
	SX127X_PIN_RST  = machine.D11
	SX127X_PIN_CS   = machine.D10
	SX127X_PIN_DIO0 = machine.D6
	SX127X_PIN_DIO1 = machine.D9
	SX127X_SPI      = machine.SPI0
)

func dioIrqHandler(machine.Pin) {
	loraRadio.HandleInterrupt()
}

func main() {
	time.Sleep(5 * time.Second)
	println("\n# TinyGo Lora RX/TX test")
	println("# ----------------------")
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	SX127X_PIN_RST.Configure(machine.PinConfig{Mode: machine.PinOutput})

	SX127X_SPI.Configure(machine.SPIConfig{Frequency: 500000, Mode: 0})

	println("main: create and start SX127x driver")
	loraRadio = sx127x.New(SX127X_SPI, SX127X_PIN_RST)
	loraRadio.SetRadioController(sx127x.NewRadioControl(SX127X_PIN_CS, SX127X_PIN_DIO0, SX127X_PIN_DIO1))

	loraRadio.Reset()
	state := loraRadio.DetectDevice()
	if !state {
		panic("main: sx127x NOT FOUND !!!")
	} else {
		println("main: sx127x found")
	}

	// Prepare for Lora Operation
	loraConf := lora.Config{
		Freq:           lora.MHz_868_1,
		Bw:             lora.Bandwidth_125_0,
		Sf:             lora.SpreadingFactor9,
		Cr:             lora.CodingRate4_7,
		HeaderType:     lora.HeaderExplicit,
		Preamble:       12,
		Iq:             lora.IQStandard,
		Crc:            lora.CRCOn,
		SyncWord:       lora.SyncPrivate,
		LoraTxPowerDBm: 20,
	}

	loraRadio.LoraConfig(loraConf)

	var count uint
	for {
		tStart := time.Now()

		println("main: Receiving Lora for 10 seconds")
		for time.Since(tStart) < 10*time.Second {
			buf, err := loraRadio.Rx(LORA_DEFAULT_RXTIMEOUT_MS)
			if err != nil {
				println("RX Error: ", err)
			} else if buf != nil {
				println("Packet Received: len=", len(buf), string(buf))
			}
		}
		println("main: End Lora RX")
		println("LORA TX size=", len(txmsg), " -> ", string(txmsg))
		err := loraRadio.Tx(txmsg, LORA_DEFAULT_TXTIMEOUT_MS)
		if err != nil {
			println("TX Error:", err)
		}
		count++
	}

}
