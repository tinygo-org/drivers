package main

// In this example, a Lora packet will be sent every 10s
// module will be in RX mode between two transmissions

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lora"
	"tinygo.org/x/drivers/sx126x"
)

const (
	LORA_DEFAULT_RXTIMEOUT_MS = 1000
	LORA_DEFAULT_TXTIMEOUT_MS = 5000
)

var (
	loraRadio *sx126x.Device
	txmsg     = []byte("Hello TinyGO")
)

func main() {
	time.Sleep(3 * time.Second)

	println("\n# TinyGo Lora RX/TX test")
	println("# ----------------------")
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Create the driver
	loraRadio = sx126x.New(spi)
	loraRadio.SetDeviceType(sx126x.DEVICE_TYPE_SX1262)

	// Create radio controller for target
	loraRadio.SetRadioController(newRadioControl())

	// Detect the device
	state := loraRadio.DetectDevice()
	if !state {
		panic("sx126x not detected.")
	}

	loraConf := lora.Config{
		Freq:           lora.MHz_868_1,
		Bw:             lora.Bandwidth_125_0,
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
		start := time.Now()

		println("main: Receiving Lora for 10 seconds")
		for time.Since(start) < 10*time.Second {
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
