//go:build pyportal || nano_rp2040 || metro_m4_airlift || arduino_mkrwifi1010 || matrixportal_m4

// +build: pyportal nano_rp2040 metro_m4_airlift arduino_mkrwifi1010 matrixportal_m4

package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/wifinina"
)

var cfg = wifinina.Config{
	// WiFi AP credentials
	Ssid:       ssid,
	Passphrase: pass,
	// Configure SPI for 8Mhz, Mode 0, MSB First
	Spi:  machine.NINA_SPI,
	Freq: 8 * 1e6,
	Sdo:  machine.NINA_SDO,
	Sdi:  machine.NINA_SDI,
	Sck:  machine.NINA_SCK,
	// Device pins
	Cs:     machine.NINA_CS,
	Ack:    machine.NINA_ACK,
	Gpio0:  machine.NINA_GPIO0,
	Resetn: machine.NINA_RESETN,
	// Watchdog (set to 0 to disable)
	WatchdogTimeo: time.Duration(20 * time.Second),
}

var netdev = wifinina.New(&cfg)
