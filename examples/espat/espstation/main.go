// This is a sensor station that uses a ESP8266 or ESP32 running on the device UART1.
// It creates a UDP connection you can use to get info to/from your computer via the microcontroller.
//
// In other words:
// Your computer <--> UART0 <--> MCU <--> UART1 <--> ESP8266
//
package main

import (
	"machine"
	"time"

	"github.com/tinygo-org/drivers/espat"
)

// access point info
const ssid = "YOURSSID"
const pass = "YOURPASS"

// IP address of the listener aka "hub". Replace with your own info.
const hubIP = "0.0.0.0"

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

	// Init esp8266/esp32
	adaptor = espat.New(uart)
	adaptor.Configure()

	// first check if connected
	if adaptor.Connected() {
		console.Write([]byte("Connected to wifi adaptor.\r\n"))
		adaptor.Echo(false)

		connectToAP()
	} else {
		console.Write([]byte("\r\n"))
		console.Write([]byte("Unable to connect to wifi adaptor.\r\n"))
		return
	}

	// now make UDP connection
	ip := espat.ParseIP(hubIP)
	raddr := &espat.UDPAddr{IP: ip, Port: 2222}
	laddr := &espat.UDPAddr{Port: 2222}

	console.Write([]byte("Dialing UDP connection...\r\n"))
	conn, _ := adaptor.DialUDP("udp", laddr, raddr)

	for {
		// send data
		console.Write([]byte("Sending data...\r\n"))
		conn.Write([]byte("hello\r\n"))

		time.Sleep(1000 * time.Millisecond)
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
