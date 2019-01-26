// This is a sensor hub that uses a ESP8266 running on the device UART1.
// It creates a UDP "server" you can use to get info to/from your computer via the microcontroller.
//
// In other words:
// Your computer <--> UART0 <--> MCU <--> UART1 <--> ESP8266 <--> INTERNET
//
package main

import (
	"machine"
	"time"

	"github.com/tinygo-org/drivers/esp8266"
)

// access point info
const ssid = "YOURSSID"
const pass = "YOURPASS"

// change these to connect to a different UART or pins for the ESP8266
var (
	uart       = machine.UART1
	tx   uint8 = machine.D10
	rx   uint8 = machine.D11

	console = machine.UART0

	adaptor *esp8266.Device
)

func main() {
	uart.Configure(machine.UARTConfig{TX: tx, RX: rx})

	// Init esp8266
	dev := esp8266.New(uart)
	adaptor = &dev
	adaptor.Configure()

	// first check if connected
	if adaptor.Connected() {
		console.Write([]byte("Connected to ESP8266.\r\n"))
		adaptor.Echo(false)

		connectToAP()
	} else {
		console.Write([]byte("\r\n"))
		console.Write([]byte("Unable to connect to esp8266.\r\n"))
		return
	}

	// now make UDP connection
	laddr := &esp8266.UDPAddr{Port: 2222}
	console.Write([]byte("Loading UDP listener...\r\n"))
	conn, _ := adaptor.ListenUDP("UDP", laddr)

	console.Write([]byte("Waiting for data...\r\n"))
	data := make([]byte, 50)
	for {
		n, _ := conn.Read(data)
		if n > 0 {
			println(string(data))
			for i := 0; i < n; i++ {
				data[i] = 0
			}
			conn.Write([]byte("hello back\r\n"))
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
	adaptor.SetWifiMode(esp8266.WifiModeClient)
	adaptor.ConnectToAP(ssid, pass, 10)
	console.Write([]byte("Connected.\r\n"))
	console.Write(adaptor.GetClientIP())
	console.Write([]byte("\r\n"))
}
