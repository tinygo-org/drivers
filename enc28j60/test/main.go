package main

import (
	"bytes"

	"tinygo.org/x/drivers/net"

	"machine"

	"tinygo.org/x/drivers/enc28j60"
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

// Arduino uno CS Pin
// var spiCS = machine.D10 // on Arduino Uno

func main() {
	// 500 bytes is usually enough for small projects
	var buff [500]byte
	// SPI Chip select pin. Can be any Digital pin
	var spiCS = machine.D53
	// Inline declarations so not used as RAM
	var ( // be sure to use the "tinygo.org/x/drivers/net" for `net` package
		MAC  = net.HardwareAddr{0xDE, 0xAD, 0xBE, 0xEF, 0xFE, 0xFF}
		MyIP = net.IP{192, 168, 1, 5} //static setup is the only one available
	)
	// Machine-specific configuration
	// use pin D0 as output
	// 8MHz SPI clk
	machine.SPI0.Configure(machine.SPIConfig{Frequency: 8e6})

	e := enc28j60.New(spiCS, machine.SPI0)

	err := e.Init(buff[:], MAC)
	if err != nil {
		println(err.Error())
	}
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led := false
	e.HTTPListenAndServe(MyIP, func(url []byte) (response []byte) {
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
