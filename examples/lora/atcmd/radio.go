package main

import (
	"errors"
)

var (
	errRadioNotFound = errors.New("radio not found")
	errRxTimeout     = errors.New("radio RX timeout")
)

type LoraRadio interface {
	Reset()
	Tx(pkt []uint8, timeoutMs uint32) error
	Rx(timeoutMs uint32) ([]uint8, error)
	SetFrequency(freq uint32)
	SetIqMode(mode uint8)
	SetCodingRate(cr uint8)
	SetBandwidth(bw uint8)
	SetCrc(enable bool)
	SetSpreadingFactor(sf uint8)
}
