// This is a sensor hub that uses a ESP8266/ESP32 running on the device UART1.
// It creates a UDP "server" you can use to get info to/from your computer via the microcontroller.
//
// In other words:
// Your computer <--> UART0 <--> MCU <--> UART1 <--> ESP8266 <--> INTERNET
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

	readyled := machine.GPIO{machine.LED}
	readyled.Configure(machine.GPIOConfig{Mode: machine.GPIO_OUTPUT})
	readyled.High()

	// first check if connected
	if adaptor.Connected() {
		console.Write([]byte("Connected to wifi adaptor.\r\n"))
		adaptor.Echo(false)

		if actAsAP {
			provideAP()
		} else {
			connectToAP()
		}
	} else {
		console.Write([]byte("\r\n"))
		console.Write([]byte("Unable to connect to wifi adaptor.\r\n"))
		return
	}

	// now make UDP connection
	laddr := &espat.UDPAddr{Port: 2222}
	console.Write([]byte("Loading UDP listener...\r\n"))
	conn, _ := adaptor.ListenUDP("UDP", laddr)

	console.Write([]byte("Waiting for data...\r\n"))
	data := make([]byte, 50)
	blink := true
	for {
		n, _ := conn.Read(data)
		if n > 0 {
			console.Write(data[:n])
			console.Write([]byte("\r\n"))
			conn.Write([]byte("hello back\r\n"))
		}
		blink = !blink
		if blink {
			readyled.High()
		} else {
			readyled.Low()
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Right now this code is never reached. Need a way to trigger it...
	console.Write([]byte("Disconnecting UDP...\r\n"))
	conn.Close()
	console.Write([]byte("Done.\r\n"))
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
