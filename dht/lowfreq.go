// +build arduino atmega1284p nrf52840 digispark nrf52 arduino_nano nrf51 atsamd21 fe310 arduino_nano33 circuitplay_express arduino_mega2560

package dht // import "tinygo.org/x/drivers/dht"

// This file provides a definition of the counter for boards with frequency lower than 2^8 ticks per millisecond (<64MHz)
type counter uint16
