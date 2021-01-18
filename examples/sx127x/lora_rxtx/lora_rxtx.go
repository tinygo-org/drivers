package main

// This example code enable continuous Wave/Preamble

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/sx127x"
)

const (
	FREQ = 868100000
)

var (
	loraRadio *sx127x.Device

	// We assume the module is connected this way:
	SX127X_PIN_RST  = machine.PB9
	SX127X_PIN_CS   = machine.PB8
	SX127X_PIN_DIO0 = machine.PA0
	SX127X_SPI      = machine.SPI0
)

func main() {
	println("\n# TinyGo Lora continuous Wave/Preamble test")
	println("# -----------------------------------------")

	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// SPI, RESET, CS configuration
	SX127X_PIN_RST.Configure(machine.PinConfig{Mode: machine.PinOutput})
	SX127X_PIN_CS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	SX127X_PIN_DIO0.Configure(machine.PinConfig{Mode: machine.PinInputPulldown}) // Checkthat
	SX127X_SPI.Configure(machine.SPIConfig{Frequency: 500000, Mode: 0})

	// Create the driver
	loraRadio = sx127x.New(SX127X_SPI, SX127X_PIN_CS, SX127X_PIN_RST)

	err := SX127X_PIN_DIO0.SetInterrupt(machine.PinRising, func(machine.Pin) {
		loraRadio.HandleInterrupt()
	})
	if err != nil {
		panic("Can't configure DIO int handler")
	}

	// Reset the radio module
	loraRadio.Reset()

	println("Try to detect sx127x radio")
	state := loraRadio.DetectDevice()
	if !state {
		panic("sx127x not detected. ")
	} else {
		println("sx127x detected")
	}

	/*
		// Prepare for Lora operation
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
				LoraTxPowerDBm: 14,
			}
			loraRadio.LoraConfig(loraConf)
	*/

	println("Start Configure Lora")
	loraRadio.SetOpModeLora()
	loraRadio.SetOpMode(sx127x.SX127X_OPMODE_SLEEP)
	loraRadio.SetTxPower(11, true)
	loraRadio.SetFrequency(FREQ)
	loraRadio.SetBandwidth(sx127x.SX127X_LORA_BW_125_0)
	loraRadio.SetCodingRate(sx127x.SX127X_LORA_CR_4_5)
	loraRadio.SetRxPayloadCrc(sx127x.SX127X_LORA_CRC_ON)
	loraRadio.SetHeaderMode(sx127x.SX127X_LORA_HEADER_EXPLICIT)
	loraRadio.SetLowDataRateOptim(sx127x.SX127X_LOW_DATARATE_OPTIM_OFF)
	println("End Configure Lora")

	for {
		loraRadio.TxLora([]byte("Hello"))
		machine.LED.Set(!machine.LED.Get())
		time.Sleep(time.Second * 2)
	}
	/*


		// Although LoraConfig has already configured most of Lora settings,
		// the following lines are still required to enable Continuous Preamble/Wave
		loraRadio.SetPacketType(sx126x.SX126X_PACKET_TYPE_LORA)
		loraRadio.SetRfFrequency(loraConf.Freq)
		loraRadio.SetModulationParams(loraConf.Sf, loraConf.Bw, loraConf.Cr, loraConf.Ldr)
		loraRadio.SetTxParams(loraConf.LoraTxPowerDBm, sx126x.SX126X_PA_RAMP_200U)

		for {
			println("2 seconds in Continuous Preamble")
			loraRadio.SetStandby()
			loraRadio.SetTxContinuousPreamble()
			time.Sleep(2 * time.Second)
			println("Continuous Preamble Stopped")

			loraRadio.SetStandby()
			time.Sleep(10 * time.Second)

			println("2 seconds in Continuous Wave")
			loraRadio.SetTxContinuousWave()
			time.Sleep(2 * time.Second)
			println(" Continuous Wave Stopped")

			loraRadio.SetStandby()
			time.Sleep(60 * time.Second)

		}
	*/
}
