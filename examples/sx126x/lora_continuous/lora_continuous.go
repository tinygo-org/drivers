package main

// This example will periodically enable Continuous "Preamble" and "Wave" modes  on 868.1 Mhz
import (
	"machine"
	"time"

	rfswitch "tinygo.org/x/drivers/examples/sx126x/rfswitch"

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
	loraRadio = sx126x.New(machine.SPI3)
	loraRadio.SetDeviceType(sx126x.DEVICE_TYPE_SX1262)

	// Create RF Switch
	var radioSwitch rfswitch.CustomSwitch
	loraRadio.SetRfSwitch(radioSwitch)

	state := loraRadio.DetectDevice()
	if !state {
		panic("sx126x not detected. ")
	}

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
