package main

// This example code demonstrates Lora RX/TX With SX127x driver
// You need to connect SPI, DIO0 and DIO1 to use.

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/sx127x"
)

const (
	LORA_DEFAULT_RXTIMEOUT_MS = 1000
	LORA_DEFAULT_TXTIMEOUT_MS = 5000

	DIO_PIN_CHANGE = machine.PinRising
)

var (
	loraRadio *sx127x.Device

	// We assume the module is connected this way:
	SX127X_PIN_RST  = machine.PB9
	SX127X_PIN_CS   = machine.PB8
	SX127X_PIN_DIO0 = machine.PA0
	SX127X_PIN_DIO1 = machine.PA1
	SX127X_SPI      = machine.SPI0

	txmsg = []byte("Hello TinyGO")

	// Prepare for Lora Operation
	loraConf = sx127x.LoraConfig{
		Freq:           868100000,
		Bw:             sx127x.SX127X_LORA_BW_125_0,
		Sf:             sx127x.SX127X_LORA_SF9,
		Cr:             sx127x.SX127X_LORA_CR_4_7,
		HeaderType:     sx127x.SX127X_LORA_HEADER_EXPLICIT,
		Preamble:       12,
		Iq:             sx127x.SX127X_LORA_IQ_STANDARD,
		Crc:            sx127x.SX127X_LORA_CRC_ON,
		SyncWord:       sx127x.SX127X_LORA_MAC_PUBLIC_SYNCWORD,
		LoraTxPowerDBm: 20,
	}
)

func main() {
	println("\n# TinyGo Driver SX127X RX/TX example")
	println("# ------------------------------------")

	println("main: configuring LED/SPI/DIO/IRQ")
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
	err := SX127X_PIN_DIO0.SetInterrupt(DIO_PIN_CHANGE, func(machine.Pin) {
		if SX127X_PIN_DIO0.Get() {
			loraRadio.HandleInterrupt()
		}
	})
	if err != nil {
		println("could not configure DIO0 pin interrupt:", err.Error())
	}

	// Setup DIO1 interrupt Handling
	err = SX127X_PIN_DIO1.SetInterrupt(DIO_PIN_CHANGE, func(machine.Pin) {
		if SX127X_PIN_DIO1.Get() {
			loraRadio.HandleInterrupt()
		}
	})
	if err != nil {
		println("could not configure DIO1 pin interrupt:", err.Error())
	}

	println("main: Configure lora modulation")
	loraRadio.LoraConfig(loraConf)

	// Get uint32 from RSSI
	rand32 := loraRadio.RandomU32()
	println("main: Get random 32bit from RSSI:", rand32)

	var count uint
	for {
		tStart := time.Now()
		machine.LED.Set(!machine.LED.Get())

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
