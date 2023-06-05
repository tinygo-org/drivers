// AT command set console running on the device UART to communicate with
// an attached LoRa device.
//
// Computer <-> UART <-> MCU <-> SPI <-> SX126x/SX127x
//
// Connect using default baudrate for this hardware, 8-N-1 with your terminal program.
// For details on the AT command set, see:
// https://files.seeedstudio.com/products/317990687/res/LoRa-E5%20AT%20Command%20Specification_V1.0%20.pdf
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/examples/lora/lorawan/common"
	"tinygo.org/x/drivers/lora"
	"tinygo.org/x/drivers/lora/lorawan"
	"tinygo.org/x/drivers/lora/lorawan/region"
)

// change these to test a different UART or pins if available
var (
	uart  = machine.Serial
	tx    = machine.UART_TX_PIN
	rx    = machine.UART_RX_PIN
	input = make([]byte, 0, 64)

	radio   lora.Radio
	session *lorawan.Session
	otaa    *lorawan.Otaa

	defaultTimeout uint32 = 1000
)

func main() {
	uart.Configure(machine.UARTConfig{TX: tx, RX: rx})

	var err error
	radio, err = common.SetupLora()
	if err != nil {
		fail(err.Error())
	}

	session = &lorawan.Session{}
	otaa = &lorawan.Otaa{}
	lorawan.UseRadio(radio)

	lorawan.UseRegionSettings(region.EU868())

	for {
		if uart.Buffered() > 0 {
			data, _ := uart.ReadByte()

			switch data {
			case 13:
				// return key
				if err := parse(input); err != nil {
					uart.Write([]byte("ERROR: "))
					uart.Write([]byte(err.Error()))
					crlf()
				}
				input = input[:0]
			default:
				// just capture the character
				input = append(input, data)
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func fail(msg string) {
	for {
		uart.Write([]byte(msg))
		crlf()

		time.Sleep(time.Minute)
	}
}
