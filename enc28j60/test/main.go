package main

import (
	"bytes"

	"machine"

	"tinygo.org/x/drivers/enc28j60"
	"tinygo.org/x/drivers/frame"
	"tinygo.org/x/drivers/net2"
)

/* Arduino Uno SPI pins:
sck:  PB5, // is D13
sdo:  PB3, // MOSI is D11
sdi:  PB4, // MISO is D12
cs:   PB2} // CS  is D10
*/

/* Arduino MEGA 2560 SPI pins as taken from tinygo library (online documentation seems to be wrong at times)
SCK: PB1 == D52
MOSI(sdo): PB2 == D51
MISO(sdi): PB3 == D50
CS: PB0 == D53
*/

// Arduino Mega 2560 CS pin
var spiCS = machine.D53

// Arduino uno CS Pin
// var spiCS = machine.D10 // on Arduino Uno

// declare as global value, can't escape RAM usage
var buff [500]byte

// Inline declarations so not used as RAM
var (
	macAddr = net2.HardwareAddr{0xDE, 0xAD, 0xBE, 0xEF, 0xFF, 0xFF}
	ipAddr  = net2.IP{192, 168, 1, 5}
)

func main() {
	frame.SDB = false
	// Machine-specific configuration
	// use pin D0 as output
	// 8MHz SPI clk
	machine.SPI0.Configure(machine.SPIConfig{Frequency: 4e6})

	e := enc28j60.New(spiCS, machine.SPI0)

	err := e.Init(buff[:], macAddr)
	if err != nil {
		println(err.Error())
	}
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led := false
	e.HTTPListenAndServe(ipAddr, func(url []byte) (response []byte) {
		if bytes.Equal(url, []byte("/led")) {
			if led {
				machine.LED.Low()
			} else {
				machine.LED.High()
			}
			led = !led
		}
		return []byte(`<h1>TinyGo Ethernet</h1><a href="led">Toggle LED</a>`)
	})
}

type handler func(url []byte) (response []byte)

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
