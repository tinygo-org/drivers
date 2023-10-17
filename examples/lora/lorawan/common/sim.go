//go:build !featherwing && !lgt92 && !stm32wlx && !sx126x

package common

import "tinygo.org/x/drivers/lora"

// do simulator setup here
func SetupLora() (lora.Radio, error) {
	return &SimLoraRadio{}, nil
}

type SimLoraRadio struct {
}

func (sr *SimLoraRadio) Reset() {
}

func (sr *SimLoraRadio) Tx(pkt []uint8, timeoutMs uint32) error {
	return nil
}

func (sr *SimLoraRadio) Rx(timeoutMs uint32) ([]uint8, error) {
	return nil, nil
}

func (sr *SimLoraRadio) SetFrequency(freq uint32)       {}
func (sr *SimLoraRadio) SetIqMode(mode uint8)           {}
func (sr *SimLoraRadio) SetCodingRate(cr uint8)         {}
func (sr *SimLoraRadio) SetBandwidth(bw uint8)          {}
func (sr *SimLoraRadio) SetCrc(enable bool)             {}
func (sr *SimLoraRadio) SetSpreadingFactor(sf uint8)    {}
func (sr *SimLoraRadio) SetHeaderType(headerType uint8) {}
func (sr *SimLoraRadio) SetPreambleLength(pLen uint16)  {}
func (sr *SimLoraRadio) SetPublicNetwork(enabled bool)  {}
func (sr *SimLoraRadio) SetSyncWord(syncWord uint16)    {}
func (sr *SimLoraRadio) SetTxPower(txPower int8)        {}
func (sr *SimLoraRadio) LoraConfig(cnf lora.Config)     {}

func FirmwareVersion() string {
	return "simulator " + CurrentVersion()
}

func Lorarx() ([]byte, error) {
	return nil, nil
}
