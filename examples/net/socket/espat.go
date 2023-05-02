//go:build challenger_rp2040

// +build: challenger_rp2040

package main

import (
	"machine"

	"tinygo.org/x/drivers/espat"
)

var cfg = espat.Config{
	// WiFi AP credentials
	Ssid:       ssid,
	Passphrase: pass,
	// UART
	Uart: machine.UART1,
	Tx:   machine.UART1_TX_PIN,
	Rx:   machine.UART1_RX_PIN,
}

var netdev = espat.New(&cfg)
