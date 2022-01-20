package main

// This example code demonstrates Lora RX/TX.
// It has been tested with bluepill and SX1276 (RFM95W board)

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/sx127x"
)

const (
	FREQ                      = 868100000
	LORA_DEFAULT_RXTIMEOUT_MS = 1000
	LORA_DEFAULT_TXTIMEOUT_MS = 5000
)

var (
	loraRadio *sx127x.Device

	// We assume the module is connected this way:
	SX127X_PIN_RST  = machine.PB9
	SX127X_PIN_CS   = machine.PB8
	SX127X_PIN_DIO0 = machine.PA0
	SX127X_SPI      = machine.SPI0

	txmsg = []byte("Hello TinyGO")
)

func main() {
	println("\n# TinyGo Lora continuous Wave/Preamble test")
	println("# -----------------------------------------")

	println("main: configuring LED/SPI/DIO/IRQ")
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	SX127X_PIN_RST.Configure(machine.PinConfig{Mode: machine.PinOutput})
	SX127X_PIN_CS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	SX127X_PIN_DIO0.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	SX127X_SPI.Configure(machine.SPIConfig{Frequency: 500000, Mode: 0})
	err := SX127X_PIN_DIO0.SetInterrupt(machine.PinRising, func(machine.Pin) { // Int handler
		loraRadio.HandleInterrupt()
	})
	if err != nil {
		panic("main: Can't configure DIO int handler")
	}

	println("main: create and start SX127x driver")
	loraRadio = sx127x.New(SX127X_SPI, SX127X_PIN_CS, SX127X_PIN_RST)
	loraRadio.Reset()
	state := loraRadio.DetectDevice()
	if !state {
		panic("main: sx127x NOT FOUND !!!")
	} else {
		println("main: sx127x found")
	}

	println("main: Configure lora modulation")
	loraRadio.SetOpModeLora()
	loraRadio.SetOpMode(sx127x.SX127X_OPMODE_SLEEP)
	loraRadio.SetTxPower(11, true)
	loraRadio.SetLoraFrequency(FREQ)
	loraRadio.SetLoraBandwidth(sx127x.SX127X_LORA_BW_125_0)
	loraRadio.SetLoraSpreadingFactor(sx127x.SX127X_LORA_SF11)
	loraRadio.SetLoraCodingRate(sx127x.SX127X_LORA_CR_4_5)
	loraRadio.SetLoraIqMode(sx127x.SX127X_LORA_IQ_STANDARD)
	loraRadio.SetLoraCrc(true)
	loraRadio.SetLoraHeaderMode(sx127x.SX127X_LORA_HEADER_EXPLICIT)
	loraRadio.SetLowDataRateOptim(sx127x.SX127X_LOW_DATARATE_OPTIM_OFF)

	// Get uint32 from RSSI
	rand32 := loraRadio.RandomU32()
	println("main: Get random 32bit from RSSI:", rand32)

	var count uint
	for {
		tStart := time.Now()
		machine.LED.Set(!machine.LED.Get())

		println("main: Receiving Lora for 10 seconds")
		for int(time.Now().Sub(tStart).Seconds()) < 60 {
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
