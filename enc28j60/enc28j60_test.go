package enc28j60_test

import (
	"bytes"
	"machine"
	"time"

	"github.com/soypat/net"

	swtch "github.com/soypat/ether-swtch"
	"tinygo.org/x/drivers/enc28j60"
)

func ExampleHTTPResponse() {
	// CS pin can be any digital ouput pin.
	var spiCS = machine.D53

	var (
		MAC  = net.HardwareAddr{0xDE, 0xAD, 0xBE, 0xEF, 0xFE, 0xFF}
		MyIP = net.IP{192, 168, 1, 5} //static setup is the only one available
	)

	// 8MHz SPI clk for older than Rev 6 boards (See Rev. B4 Silicon Errata)
	machine.SPI0.Configure(machine.SPIConfig{Frequency: 8e6})

	e := enc28j60.New(spiCS, machine.SPI0)
	err := e.Init(MAC)
	if err != nil {
		println(err.Error())
	}
	const okHeader = "HTTP/1.0 200 OK\r\nContent-Type: text/html\r\nPragma: no-cache\r\n\r\n"

	timeout := time.Second * 1 // timeouts don't work well on AVR boards yet
	swtch.HTTPListenAndServe(e, MAC, MyIP, timeout, func(URL []byte) (response []byte) {
		return []byte(okHeader + "Hello world!")
	}, func(e error) {}) // Will ignore errors.
}

func ExampleBlinky() {
	// CS pin can be any digital ouput pin.
	var spiCS = machine.D53

	var (
		MAC  = net.HardwareAddr{0xDE, 0xAD, 0xBE, 0xEF, 0xFE, 0xFF}
		MyIP = net.IP{192, 168, 1, 5} //static setup is the only one available
	)

	// 8MHz SPI clk for older than Rev 6 boards (See Rev. B4 Silicon Errata)
	machine.SPI0.Configure(machine.SPIConfig{Frequency: 8e6})
	machine.LED.Configure(machine.PinConfig{Mode: machine.PinOutput})

	e := enc28j60.New(spiCS, machine.SPI0)
	err := e.Init(MAC)
	if err != nil {
		println(err.Error())
	}
	const okHeader = "HTTP/1.0 200 OK\r\nContent-Type: text/html\r\nPragma: no-cache\r\n\r\n"

	timeout := time.Second * 1
	swtch.HTTPListenAndServe(e, MAC, MyIP, timeout, func(URL []byte) (response []byte) {
		// Warning: This will leave you at URL "/led". Refreshing the page will trigger the toggle!
		// Use HTTP redirects to fix this.
		if bytes.Equal(URL, []byte("/led")) {
			machine.LED.Set(!machine.LED.Get())
		}
		return []byte(okHeader + `<h1>TinyGo Ethernet</h1><a href="led">Toggle LED</a>`)
	}, func(e error) {}) // Will ignore errors.
}
