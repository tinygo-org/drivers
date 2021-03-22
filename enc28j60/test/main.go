package main

import (
	"machine"

	"tinygo.org/x/drivers/enc28j60"
)

/* Arduino MEGA 2560 SPI pins as taken from tinygo library (online documentation seems to be wrong at times)
SCK: PB1 == D52
MISO(sdo): PB2 == D51
MOSI(sdi): PB3 == D50
CS: PB0 == D53
	sdo:  PB2,
	sdi:  PB3,
*/

// Arduino Mega 2560 CS pin
var spiCS = machine.D10

// Arduino uno CS Pin
// var spiCS = machine.D10

// declare as global value, can't escape RAM usage
var buff [1000]byte

func main() {
	// Inline declarations so not used as RAM
	var (
	// gateway or router address
	// gwAddr = net.IP{192, 168, 1, 1}
	// // IP address of ENC28J60
	// ipAddr = net.IP{192, 168, 1, 5}
	// // Hardware address of ENC28J60
	// macAddr = net.HardwareAddr{0xfe, 0xfe, 0xfe, 0x22, 0x22, 0x22}
	// // network mask
	// netmask = net.IPMask{255, 255, 255, 0}
	)
	enc28j60.SDB = true
	// Machine-specific configuration
	spiCS.Configure(machine.PinConfig{Mode: machine.PinOutput})
	spiCS.High() // prevent SPI glitches
	machine.SPI0.Configure(machine.SPIConfig{Mode: machine.Mode0, LSBFirst: false})

	enc28j60.TestSPI(spiCS, machine.SPI0)
}

func printError(err error) {
	if enc28j60.SDB {
		println(err.Error())
	} else {
		println("error #", err.(enc28j60.ErrorCode))
	}
}

func byteSliceToHex(b []byte) []byte {
	o := make([]byte, len(b)*2)
	for i := 0; i < len(b); i++ {
		aux := byteToHex(b[i])
		o[i*2] = aux[0]
		o[i*2+1] = aux[1]
	}
	return o
}

func byteToHex(b byte) []byte {
	var res [2]byte
	res[0], res[1] = (b>>4)+'0', (b&0b0000_1111)+'0'
	if (b >> 4) > 9 {
		res[0] = (b >> 4) + 'A' - 10
	}
	if (b & 0b0000_1111) > 9 {
		res[1] = (b & 0b0000_1111) + 'A' - 10
	}
	return res[:]
}
