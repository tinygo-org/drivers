//go:build !featherwing && !gnse && !lorae5 && !nucleowl55jc

package main

// do simulator setup here
func setupLora() (LoraRadio, error) {
	return &SimLoraRadio{}, nil
}

type SimLoraRadio struct {
}

func (sr *SimLoraRadio) Reset() {
}

func (sr *SimLoraRadio) LoraTx(pkt []uint8, timeoutMs uint32) error {
	return nil
}

func (sr *SimLoraRadio) LoraRx(timeoutMs uint32) ([]uint8, error) {
	return nil, nil
}

func (sr *SimLoraRadio) SetLoraFrequency(freq uint32)    {}
func (sr *SimLoraRadio) SetLoraIqMode(mode uint8)        {}
func (sr *SimLoraRadio) SetLoraCodingRate(cr uint8)      {}
func (sr *SimLoraRadio) SetLoraBandwidth(bw uint8)       {}
func (sr *SimLoraRadio) SetLoraCrc(enable bool)          {}
func (sr *SimLoraRadio) SetLoraSpreadingFactor(sf uint8) {}

func firmwareVersion() string {
	return "simulator " + currentVersion()
}

func lorarx() ([]byte, error) {
	return nil, nil
}
