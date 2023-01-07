//go:build gnse || lorae5 || nucleowl55jc

package main

import (
	"device/stm32"
	"machine"
	"runtime/interrupt"

	rfswitch "tinygo.org/x/drivers/examples/sx126x/rfswitch"

	"tinygo.org/x/drivers/lora"
	"tinygo.org/x/drivers/sx126x"
)

const FREQ = 868100000

var (
	loraRadio *sx126x.Device
)

// do sx126x setup here
func setupLora() (LoraRadio, error) {
	loraRadio = sx126x.New(machine.SPI3)
	loraRadio.SetDeviceType(sx126x.DEVICE_TYPE_SX1262)

	// Create RF Switch
	var radioSwitch rfswitch.CustomSwitch
	loraRadio.SetRfSwitch(radioSwitch)

	if state := loraRadio.DetectDevice(); !state {
		return nil, errRadioNotFound
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

	return loraRadio, nil
}

// radioIntHandler will take care of radio interrupts
func radioIntHandler(intr interrupt.Interrupt) {
	loraRadio.HandleInterrupt()
}

func firmwareVersion() string {
	return "sx126x"
}

func lorarx() ([]byte, error) {
	return loraRadio.LoraRx(LORA_DEFAULT_RXTIMEOUT_MS)
}
