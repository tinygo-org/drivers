package main

// In this example, a Lora packet will be sent every 10s
// module will be in RX mode between two transmissions

import (
	"device/stm32"
	"machine"
	"runtime/interrupt"
	"time"

	rfswitch "tinygo.org/x/drivers/examples/sx126x/rfswitch"

	"tinygo.org/x/drivers/lora"
	"tinygo.org/x/drivers/sx126x"
)

const FREQ = 868100000

const (
	LORA_DEFAULT_RXTIMEOUT_MS = 1000
	LORA_DEFAULT_TXTIMEOUT_MS = 5000
)

var (
	loraRadio *sx126x.Device
	txmsg     = []byte("Hello TinyGO")
)

// radioIntHandler will take care of radio interrupts
func radioIntHandler(intr interrupt.Interrupt) {
	loraRadio.HandleInterrupt()
}

func main() {
	println("\n# TinyGo Lora RX/TX test")
	println("# ----------------------")
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Create the driver
	loraRadio = sx126x.New(machine.SPI3)
	loraRadio.SetDeviceType(sx126x.DEVICE_TYPE_SX1262)

	// Create RF Switch
	var radioSwitch rfswitch.CustomSwitch
	loraRadio.SetRfSwitch(radioSwitch)

	// Detect the device
	state := loraRadio.DetectDevice()
	if !state {
		panic("sx126x not detected.")
	}

	// Add interrupt handler for Radio IRQs
	intr := interrupt.New(stm32.IRQ_Radio_IRQ_Busy, radioIntHandler)
	intr.Enable()

	loraConf := lora.Config{
		Freq:           FREQ,
		Bw:             lora.Bandwidth_500_0,
		Sf:             lora.SpreadingFactor9,
		Cr:             lora.CodingRate4_7,
		HeaderType:     lora.HeaderExplicit,
		Preamble:       12,
		Ldr:            lora.LowDataRateOptimizeOff,
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
