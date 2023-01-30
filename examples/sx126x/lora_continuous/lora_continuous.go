package main

// This example will periodically enable Continuous "Preamble" and "Wave" modes  on 868.1 Mhz
import (
	"machine"
	"time"

	"tinygo.org/x/drivers/lora"
	"tinygo.org/x/drivers/sx126x"
)

const FREQ = 868100000

var (
	loraRadio *sx126x.Device
)

func main() {
	println("\n# TinyGo Lora continuous Wave/Preamble test")
	println("# -----------------------------------------")

	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// Create the driver
	loraRadio = sx126x.New(spi)
	loraRadio.SetDeviceType(sx126x.DEVICE_TYPE_SX1262)

	// Create radio controller for target
	loraRadio.SetRadioController(newRadioControl())

	state := loraRadio.DetectDevice()
	if !state {
		panic("sx126x not detected. ")
	}

	// Prepare for Lora operation
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
}
