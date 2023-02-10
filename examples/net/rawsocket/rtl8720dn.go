//go:build wioterminal

// +build: wioterminal

package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/rtl8720dn"
)

var cfg = rtl8720dn.Config{
	// WiFi AP credentials
	Ssid:       ssid,
	Passphrase: pass,
	// Device
	En: machine.RTL8720D_CHIP_PU,
	// UART
	Uart:     machine.UART3,
	Tx:       machine.PB24,
	Rx:       machine.PC24,
	Baudrate: 614400,
	// Watchdog (set to 0 to disable)
	WatchdogTimeo: time.Duration(20 * time.Second),
}

var dev = rtl8720dn.New(&cfg)

func NetConnect() error {
	return dev.NetConnect()
}

func NetDisconnect() {
	dev.NetDisconnect()
}
