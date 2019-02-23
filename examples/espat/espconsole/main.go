// This is a console to a ESP8266/ESP32 running on the device UART1.
// Allows you to type AT commands from your computer via the microcontroller.
//
// In other words:
// Your computer <--> UART0 <--> MCU <--> UART1 <--> ESP8266 <--> INTERNET
//
// More information on the Espressif AT command set at:
// https://www.espressif.com/sites/default/files/documentation/4a-esp8266_at_instruction_set_en.pdf
//
package main

import (
	"machine"
	"time"

	"github.com/tinygo-org/drivers/espat"
)

// change actAsAP to true to act as an access point instead of connecting to one.
const actAsAP = false

// access point info
const ssid = "YOURSSID"
const pass = "YOURPASS"

// change these to connect to a different UART or pins for the ESP8266/ESP32
var (
	uart       = machine.UART1
	tx   uint8 = machine.D10
	rx   uint8 = machine.D11

	console = machine.UART0

	adaptor *espat.Device
)

func main() {
	uart.Configure(machine.UARTConfig{TX: tx, RX: rx})

	// Init esp8266
	adaptor = espat.New(uart)
	adaptor.Configure()

	// first check if connected
	if adaptor.Connected() {
		adaptor.Echo(false)
		console.Write([]byte("\r\n"))
		console.Write([]byte("ESP-AT console enabled.\r\n"))
		console.Write([]byte("Firmware version:\r\n"))
		console.Write(adaptor.Version())
		console.Write([]byte("\r\n"))

		if actAsAP {
			provideAP()
		} else {
			connectToAP()
		}

		console.Write([]byte("Type an AT command then press enter:\r\n"))
		prompt()
	} else {
		console.Write([]byte("\r\n"))
		console.Write([]byte("Unable to connect to wifi adaptor.\r\n"))
		return
	}

	input := make([]byte, 64)
	i := 0
	for {
		if console.Buffered() > 0 {
			data, _ := console.ReadByte()

			switch data {
			case 13:
				// return key
				console.Write([]byte("\r\n"))

				// send command to ESP8266
				input[i] = byte('\r')
				input[i+1] = byte('\n')
				adaptor.Write(input[:i+2])

				// give the ESP8266 a chance to respond.
				time.Sleep(10 * time.Millisecond)

				// display response
				console.Write(adaptor.Response())

				// prompt
				prompt()

				i = 0
				continue
			default:
				// just echo the character
				console.WriteByte(data)
				input[i] = data
				i++
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func prompt() {
	console.Write([]byte("ESPAT>"))
}

// connect to access point
func connectToAP() {
	console.Write([]byte("Connecting to wifi network...\r\n"))
	adaptor.SetWifiMode(espat.WifiModeClient)
	adaptor.ConnectToAP(ssid, pass, 10)
	console.Write([]byte("Connected.\r\n"))
	console.Write([]byte(adaptor.GetClientIP()))
	console.Write([]byte("\r\n"))
}

// provide access point
func provideAP() {
	console.Write([]byte("Starting wifi network as access point '"))
	console.Write([]byte(ssid))
	console.Write([]byte("'...\r\n"))
	adaptor.SetWifiMode(espat.WifiModeAP)
	adaptor.SetAPConfig(ssid, pass, 7, espat.WifiAPSecurityWPA2_PSK)
	console.Write([]byte("Ready.\r\n"))
	console.Write([]byte(adaptor.GetAPIP()))
	console.Write([]byte("\r\n"))
}
