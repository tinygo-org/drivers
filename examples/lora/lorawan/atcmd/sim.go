//go:build !featherwing && !gnse && !lorae5 && !nucleowl55jc

package main

import "tinygo.org/x/drivers/lora"

// do simulator setup here
func setupLora() (lora.Radio, error) {
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

func (sr *SimLoraRadio) SetFrequency(freq uint32)    {}
func (sr *SimLoraRadio) SetIqMode(mode uint8)        {}
func (sr *SimLoraRadio) SetCodingRate(cr uint8)      {}
func (sr *SimLoraRadio) SetBandwidth(bw uint8)       {}
func (sr *SimLoraRadio) SetCrc(enable bool)          {}
func (sr *SimLoraRadio) SetSpreadingFactor(sf uint8) {}

func firmwareVersion() string {
	return "simulator " + currentVersion()
}

func lorarx() ([]byte, error) {
	return nil, nil
}
