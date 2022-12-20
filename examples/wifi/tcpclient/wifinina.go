//go:build wifinina

package main

import (
	"machine"
	"tinygo.org/x/drivers/wifinina"
)

var (
	// default interface for the Arduino Nano33 IoT.
	spi = machine.NINA_SPI

	// ESP32/ESP8266 chip that has the WIFININA firmware flashed on it
	adaptor *wifinina.Device
)

func initAdaptor() *wifinina.Device {
	// Configure SPI for 8Mhz, Mode 0, MSB First
	spi.Configure(machine.SPIConfig{
		Frequency: 8 * 1e6,
		SDO:       machine.NINA_SDO,
		SDI:       machine.NINA_SDI,
		SCK:       machine.NINA_SCK,
	})

	adaptor = wifinina.New(spi,
		machine.NINA_CS,
		machine.NINA_ACK,
		machine.NINA_GPIO0,
		machine.NINA_RESETN)
	adaptor.Configure()

	return adaptor
}
