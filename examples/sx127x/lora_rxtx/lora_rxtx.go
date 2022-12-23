package main

// This example code demonstrates Lora RX/TX With SX127x driver
// You need to connect SPI, DIO0 and DIO1 to use.

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lora"
	"tinygo.org/x/drivers/sx127x"
)

const FREQ = 868100000

const (
	LORA_DEFAULT_RXTIMEOUT_MS = 1000
	LORA_DEFAULT_TXTIMEOUT_MS = 5000
)

var (
	loraRadio *sx127x.Device
	txmsg     = []byte("Hello TinyGO")

	// We assume the module is connected this way:
	SX127X_PIN_RST  = machine.PB9
	SX127X_PIN_CS   = machine.PB8
	SX127X_PIN_DIO0 = machine.PA0
	SX127X_PIN_DIO1 = machine.PA1
	SX127X_SPI      = machine.SPI0
)

func dioIrqHandler(machine.Pin) {
	loraRadio.HandleInterrupt()
}

func main() {
	println("\n# TinyGo Lora RX/TX test")
	println("# ----------------------")
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	SX127X_PIN_RST.Configure(machine.PinConfig{Mode: machine.PinOutput})
	SX127X_PIN_CS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	SX127X_PIN_DIO0.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	SX127X_PIN_DIO1.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	SX127X_SPI.Configure(machine.SPIConfig{Frequency: 500000, Mode: 0})

	println("main: create and start SX127x driver")
	loraRadio = sx127x.New(SX127X_SPI, SX127X_PIN_CS, SX127X_PIN_RST)
	loraRadio.Reset()
	state := loraRadio.DetectDevice()
	if !state {
		panic("main: sx127x NOT FOUND !!!")
	} else {
		println("main: sx127x found")
	}

	// Setup DIO0 interrupt Handling
	if err := SX127X_PIN_DIO0.SetInterrupt(machine.PinRising, dioIrqHandler); err != nil {
		println("could not configure DIO0 pin interrupt:", err.Error())
	}

	// Setup DIO1 interrupt Handling
	if err := SX127X_PIN_DIO1.SetInterrupt(machine.PinRising, dioIrqHandler); err != nil {
		println("could not configure DIO1 pin interrupt:", err.Error())
	}

	// Prepare for Lora Operation
	loraConf := lora.Config{
		Freq:           FREQ,
		Bw:             lora.Bandwidth_500_0,
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
		for int(time.Now().Sub(tStart).Seconds()) < 10 {
			buf, err := loraRadio.LoraRx(LORA_DEFAULT_RXTIMEOUT_MS)
			if err != nil {
				println("RX Error: ", err)
			} else if buf != nil {
				println("Packet Received: len=", len(buf), string(buf))
			}
		}
		println("main: End Lora RX")
		println("LORA TX size=", len(txmsg), " -> ", string(txmsg))
		err := loraRadio.LoraTx(txmsg, LORA_DEFAULT_TXTIMEOUT_MS)
		if err != nil {
			println("TX Error:", err)
		}
		count++
	}

}
