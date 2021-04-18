package enc28j60_test

import (
	"machine"

	"tinygo.org/x/drivers/enc28j60"
	"tinygo.org/x/drivers/net2"
)

func ExampleEthernetChatter() {
	// best declared as a global variable for tinygo application
	var buff [1000]byte
	// Machine-specific configuration
	// use pin D0 as output
	// 8MHz SPI clk
	machine.SPI0.Configure(machine.SPIConfig{Frequency: 8e6})
	e := enc28j60.New(machine.D10, machine.SPI0)
	MAC := net2.HardwareAddr{0xfe, 0xfe, 0xfe, 0x22, 0x22, 0x22}
	err := e.Init(buff[:], MAC)
	if err != nil {
		println(err.Error())
	}

	// TODO new example with new API
}
