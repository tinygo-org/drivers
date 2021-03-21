package main

// In this example, a Lora packet will be sent every 10s
// module will be in RX mode between two transmissions

import (
	"device/stm32"
	"machine"
	"runtime/interrupt"
	"time"

	rfswitch "tinygo.org/x/drivers/examples/sx126x/rfswitch"

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

	loraConf := sx126x.LoraConfig{
		Freq:           FREQ,
		Bw:             sx126x.SX126X_LORA_BW_500_0,
		Sf:             sx126x.SX126X_LORA_SF9,
		Cr:             sx126x.SX126X_LORA_CR_4_7,
		HeaderType:     sx126x.SX126X_LORA_HEADER_EXPLICIT,
		Preamble:       12,
		Ldr:            sx126x.SX126X_LORA_LOW_DATA_RATE_OPTIMIZE_OFF,
		Iq:             sx126x.SX126X_LORA_IQ_STANDARD,
		Crc:            sx126x.SX126X_LORA_CRC_ON,
		SyncWord:       sx126x.SX126X_LORA_MAC_PRIVATE_SYNCWORD,
		LoraTxPowerDBm: 20,
	}

	loraRadio.LoraConfig(loraConf)

	var count uint
	for {
		tStart := time.Now()

		// Blocking RX for LORA_DEFAULT_RXTIMEOUT_MS
		println("Start Lora RX for 10 sec")
		for int(time.Now().Sub(tStart).Seconds()) < 10 {
			buf, err := loraRadio.LoraRx(LORA_DEFAULT_RXTIMEOUT_MS)

			if err != nil {
				println("RX Error: ", err)
			} else if buf != nil {
				println("Packet Received: len=", len(buf), string(buf))
			}
		}
		println("END Lora RX")

		println("LORA TX size=", len(txmsg))
		err := loraRadio.LoraTx(txmsg, LORA_DEFAULT_TXTIMEOUT_MS)
		if err != nil {
			println("TX Error:", err)
		}
		count++
	}

}
