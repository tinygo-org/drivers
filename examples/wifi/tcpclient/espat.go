//go:build espat

package main

import (
	"machine"
	"tinygo.org/x/drivers/espat"
)

// these are the default pins for the Arduino Nano33 IoT.
// change these to connect to a different UART or pins for the ESP8266/ESP32
var (
	uart = machine.UART1
	tx   = machine.PA22
	rx   = machine.PA23

	adaptor *espat.Device
)

func initAdaptor() *espat.Device {
	uart.Configure(machine.UARTConfig{TX: tx, RX: rx})

	adaptor = espat.New(uart)
	adaptor.Configure()

	return adaptor
}
