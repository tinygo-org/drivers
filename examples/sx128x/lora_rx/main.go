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

type LoRaConfig struct {
	Frequency       uint32
	Power           int8
	RadioRamp       sx128x.RadioRampTime
	RegulatorMode   sx128x.RegulatorMode
	SpreadingFactor sx128x.LoRaSpreadingFactor
	Bandwidth       sx128x.LoRaBandwidth
	CodingRate      sx128x.LoRaCodingRate
	PreambleLength  uint32
	HeaderType      sx128x.LoRaHeaderType
	CrcType         sx128x.LoRaCrcType
	IqType          sx128x.LoRaIqType
	SyncWord        uint16
}

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

	loraConfig := LoRaConfig{
		Frequency:     2400000000,              // 2.4Ghz
		Power:         13,                      // dBm
		RadioRamp:     sx128x.RADIO_RAMP_02_US, // 2 microsecond ramp time
		RegulatorMode: sx128x.REGULATOR_DC_DC,

		SpreadingFactor: sx128x.LORA_SF_9,
		Bandwidth:       sx128x.LORA_BW_1600,
		CodingRate:      sx128x.LORA_CR_4_7,
		PreambleLength:  12,
		HeaderType:      sx128x.LORA_HEADER_EXPLICIT,
		CrcType:         sx128x.LORA_CRC_DISABLE,
		IqType:          sx128x.LORA_IQ_STD,
		SyncWord:        0x1424, // the default private sync word
	}

	radio.WaitWhileBusy(time.Second)
	SetupLoRa(radio, loraConfig)

	for {
		data, err := Rx(radio, 1000) // 1 second timeout
		if err != nil {
			println("failed to receive:", err)
		} else {
			println("received:", string(data))
		}
	}
}

func SetupLoRa(radio *sx128x.Device, config LoRaConfig) {
	radio.SetStandby(sx128x.STANDBY_RC)
	radio.SetPacketType(sx128x.PACKET_TYPE_LORA)
	radio.SetRegulatorMode(config.RegulatorMode)

	radio.SetRfFrequency(config.Frequency)
	radio.SetModulationParamsLoRa(config.SpreadingFactor, config.Bandwidth, config.CodingRate)

	// section 14.4.1 shows required register settings for setting up LoRa operations.
	if config.SpreadingFactor == sx128x.LORA_SF_5 || config.SpreadingFactor == sx128x.LORA_SF_6 {
		radio.WriteRegister(0x925, []byte{0x1E})
	} else if config.SpreadingFactor == sx128x.LORA_SF_7 || config.SpreadingFactor == sx128x.LORA_SF_8 {
		radio.WriteRegister(0x925, []byte{0x37})
	} else {
		radio.WriteRegister(0x925, []byte{0x32})
	}
	radio.WriteRegister(0x93C, []byte{0x01})

	radio.SetTxParams(config.Power, config.RadioRamp)
	radio.SetPacketParamsLoRa(config.PreambleLength, config.HeaderType, 0xFF, config.CrcType, config.IqType)
	radio.WriteRegister(sx128x.REG_LORA_SYNC_WORD_MSB, []byte{byte(config.SyncWord >> 8), byte(config.SyncWord & 0xFF)})
}

func Rx(radio *sx128x.Device, timeout uint16) ([]byte, error) {
	radio.SetStandby(sx128x.STANDBY_RC)
	radio.SetBufferBaseAddress(0, 0)
	radio.SetDioIrqParams(sx128x.IRQ_RX_DONE_MASK|sx128x.IRQ_RX_TX_TIMEOUT_MASK, sx128x.IRQ_RX_DONE_MASK|sx128x.IRQ_RX_TX_TIMEOUT_MASK, 0x00, 0x00)
	radio.ClearIrqStatus(sx128x.IRQ_ALL_MASK)
	radio.SetRx(sx128x.PERIOD_BASE_1_MS, timeout)
	// busy wait for IRQ indication
	for dio1Pin.Get() == false {
		runtime.Gosched()
	}
	irqStatus, _ := radio.GetIrqStatus()
	if irqStatus&sx128x.IRQ_RX_DONE_MASK != 0 {
		payloadLength, bufferOffset, err := radio.GetRxBufferStatus()
		if err != nil {
			return nil, err
		}
		data, err := radio.ReadBuffer(bufferOffset, payloadLength)
		return data, nil
	} else if irqStatus&sx128x.IRQ_RX_TX_TIMEOUT_MASK != 0 {
		return nil, errors.New("rx timeout")
	}
	return nil, errors.New("unexpected IRQ status")
}
