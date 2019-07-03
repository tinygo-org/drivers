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

	"tinygo.org/x/drivers/espat"
	"tinygo.org/x/drivers/espat/net"
)

// change actAsAP to true to act as an access point instead of connecting to one.
const actAsAP = false

// access point info
const ssid = "YOURSSID"
const pass = "YOURPASS"

// change these to connect to a different UART or pins for the ESP8266/ESP32
var (
	uart = machine.UART1
	tx   = machine.D10
	rx   = machine.D11

	adaptor *espat.Device
)

func main() {
	uart.Configure(machine.UARTConfig{TX: tx, RX: rx})

	// Init esp8266
	adaptor = espat.New(uart)
	adaptor.Configure()

	readyled := machine.LED
	readyled.Configure(machine.PinConfig{Mode: machine.PinOutput})
	readyled.High()

	// first check if connected
	if adaptor.Connected() {
		println("Connected to wifi adaptor.")
		adaptor.Echo(false)

		if actAsAP {
			provideAP()
		} else {
			connectToAP()
		}
	} else {
		println("Unable to connect to wifi adaptor.")
		return
	}

	// now make UDP connection
	laddr := &net.UDPAddr{Port: 2222}
	println("Loading UDP listener...")
	conn, _ := net.ListenUDP("UDP", laddr)

	println("Waiting for data...")
	data := make([]byte, 50)
	blink := true
	for {
		n, _ := conn.Read(data)
		if n > 0 {
			println(string(data[:n]))
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
	println("Disconnecting UDP...")
	conn.Close()
	println("Done.")
}

// connect to access point
func connectToAP() {
	println("Connecting to wifi network...")
	adaptor.SetWifiMode(espat.WifiModeClient)
	adaptor.ConnectToAP(ssid, pass, 10)
	println("Connected.")
	println(adaptor.GetClientIP())
}

// provide access point
func provideAP() {
	println("Starting wifi network as access point:")
	println(ssid)
	adaptor.SetWifiMode(espat.WifiModeAP)
	adaptor.SetAPConfig(ssid, pass, 7, espat.WifiAPSecurityWPA2_PSK)
	println("Ready.")
	println(adaptor.GetAPIP())
}
