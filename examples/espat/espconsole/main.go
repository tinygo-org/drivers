// This is a console to a ESP8266/ESP32 running on the device UART1.
// Allows you to type AT commands from your computer via the microcontroller.
//
// In other words:
// Your computer <--> UART0 <--> MCU <--> UART1 <--> ESP8266 <--> INTERNET
//
// More information on the Espressif AT command set at:
// https://www.espressif.com/sites/default/files/documentation/4a-esp8266_at_instruction_set_en.pdf
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/espat"
)

// change actAsAP to true to act as an access point instead of connecting to one.
const actAsAP = false

var (
	// access point info
	ssid string
	pass string
)

// these are the default pins for the Arduino Nano33 IoT.
// change these to connect to a different UART or pins for the ESP8266/ESP32
var (
	uart = machine.UART1
	tx   = machine.PA22
	rx   = machine.PA23

	console = machine.Serial

	adaptor *espat.Device
)

func main() {
	uart.Configure(machine.UARTConfig{TX: tx, RX: rx})

	// Init esp8266
	adaptor = espat.New(uart)
	adaptor.Configure()

	// first check if connected
	if connectToESP() {
		println("Connected to wifi adaptor.")
		adaptor.Echo(false)

		connectToAP()
	} else {
		println("")
		failMessage("Unable to connect to wifi adaptor.")
		return
	}

	println("Type an AT command then press enter:")
	prompt()

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

				// display response
				r, _ := adaptor.Response(500)
				console.Write(r)

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
	print("ESPAT>")
}

// connect to ESP8266/ESP32
func connectToESP() bool {
	for i := 0; i < 5; i++ {
		println("Connecting to wifi adaptor...")
		if adaptor.Connected() {
			return true
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

// connect to access point
func connectToAP() {
	println("Connecting to wifi network '" + ssid + "'")

	if err := adaptor.ConnectToAccessPoint(ssid, pass, 10*time.Second); err != nil {
		failMessage(err.Error())
	}

	println("Connected.")
	ip, err := adaptor.GetClientIP()
	if err != nil {
		failMessage(err.Error())
	}

	println(ip)
}

// provide access point
func provideAP() {
	println("Starting wifi network as access point '" + ssid + "'...")
	adaptor.SetWifiMode(espat.WifiModeAP)
	adaptor.SetAPConfig(ssid, pass, 7, espat.WifiAPSecurityWPA2_PSK)
	println("Ready.")
	ip, _ := adaptor.GetAPIP()
	println(ip)
}

func failMessage(msg string) {
	for {
		println(msg)
		time.Sleep(1 * time.Second)
	}
}
