package main

import (
	"errors"
	"machine"
	"runtime"
	"time"

	"tinygo.org/x/drivers/sx128x"
)

var (
	// pin mapping specific to the lilygo t3s3, change as needed for your board
	sdoPin   = machine.GPIO6
	sdiPin   = machine.GPIO3
	sckPin   = machine.GPIO5
	nssPin   = machine.GPIO7
	busyPin  = machine.GPIO36
	resetPin = machine.GPIO8
	dio1Pin  = machine.GPIO9
)

func setupPins() {
	nssPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	nssPin.Set(true)

	resetPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	resetPin.Set(true)

	busyPin.Configure(machine.PinConfig{Mode: machine.PinInput})
	dio1Pin.Configure(machine.PinConfig{Mode: machine.PinInput})
}

func main() {
	setupPins()

	spi := machine.SPI0
	spi.Configure(machine.SPIConfig{
		Mode:      0,
		Frequency: 8 * 1e6,
		SDO:       sdoPin,
		SDI:       sdiPin,
		SCK:       sckPin,
	})

	radio := sx128x.New(
		spi,
		nssPin,
		resetPin,
		busyPin,
	)

	radio.WaitWhileBusy(time.Second)
	SetupLora(radio)

	for {
		Tx(radio, []byte("Hello, world!"))
		time.Sleep(1 * time.Second)
	}
}

func SetupLora(radio *sx128x.Device) {
	radio.SetStandby(sx128x.STANDBY_RC)
	radio.SetPacketType(sx128x.PACKET_TYPE_LORA)
	radio.SetRegulatorMode(sx128x.REGULATOR_DC_DC)

	radio.SetRfFrequency(2400000000) // 2.4Ghz
	radio.SetModulationParamsLoRa(sx128x.LORA_SF_9, sx128x.LORA_BW_1600, sx128x.LORA_CR_4_7)

	// section 14.4.1 shows required register setting for setting up LoRa operations. These depend on the chosen spreading factor.
	radio.WriteRegister(0x925, []byte{0x32})
	radio.WriteRegister(0x93C, []byte{0x01})

	radio.SetTxParams(13, sx128x.RADIO_RAMP_02_US)
	radio.SetPacketParamsLoRa(12, sx128x.LORA_HEADER_EXPLICIT, 0xFF, sx128x.LORA_CRC_DISABLE, sx128x.LORA_IQ_STD)
	radio.WriteRegister(sx128x.REG_LORA_SYNC_WORD_MSB, []byte{0x14, 0x24}) // full sync word is 0x1424

}

func Tx(radio *sx128x.Device, data []byte) error {
	if len(data) > 255 {
		return errors.New("data length exceeds maximum of 255 bytes")
	}
	radio.SetStandby(sx128x.STANDBY_RC)
	radio.SetPacketParamsLoRa(12, sx128x.LORA_HEADER_EXPLICIT, uint8(len(data)&0xFF), sx128x.LORA_CRC_DISABLE, sx128x.LORA_IQ_STD)
	radio.SetBufferBaseAddress(0, 0)
	radio.WriteBuffer(0, data)
	radio.SetDioIrqParams(sx128x.IRQ_TX_DONE_MASK|sx128x.IRQ_RX_TX_TIMEOUT_MASK, sx128x.IRQ_TX_DONE_MASK|sx128x.IRQ_RX_TX_TIMEOUT_MASK, 0x00, 0x00)
	radio.ClearIrqStatus(sx128x.IRQ_ALL_MASK)
	radio.SetTx(sx128x.PERIOD_BASE_4_MS, 250) // 4ms * 250 = 1s
	// busy wait for IRQ indication
	for dio1Pin.Get() == false {
		runtime.Gosched()
	}
	return nil
}
