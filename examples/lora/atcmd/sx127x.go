//go:build featherwing

package main

import (
	"strconv"

	"machine"

	"tinygo.org/x/drivers/lora"
	"tinygo.org/x/drivers/sx127x"
)

const FREQ = 868100000

var (
	// We assume LoRa Featherwing module is connected to PyBadge:
	rstPin  = machine.D11
	csPin   = machine.D10
	dio0Pin = machine.D6
	dio1Pin = machine.D9
	spi = machine.SPI0
	loraRadio *sx127x.Device
)

// do sx127x setup here
func setupLora() (LoraRadio, error) {
	rstPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	csPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	dio0Pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	dio1Pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	spi.Configure(machine.SPIConfig{Frequency: 500000, Mode: 0})

	loraRadio = sx127x.New(spi, csPin, rstPin)
	loraRadio.Reset()

	if state := loraRadio.DetectDevice(); !state {
		return nil, errRadioNotFound
	}

	// Setup DIO0 interrupt Handling
	if err := dio0Pin.SetInterrupt(machine.PinRising, dioIrqHandler); err != nil {
		println("could not configure DIO0 pin interrupt:", err.Error())
	}

	// Setup DIO1 interrupt Handling
	if err := dio1Pin.SetInterrupt(machine.PinRising, dioIrqHandler); err != nil {
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

	return loraRadio, nil	
}


func dioIrqHandler(machine.Pin) {
	loraRadio.HandleInterrupt()
}

func firmwareVersion() string {
	v := loraRadio.GetVersion()
	return "sx127x v"+strconv.Itoa(int(v))
}
