// This is a sensor hub that uses a ESP8266/ESP32 running on the device UART1.
// It creates a UDP "server" you can use to get info to/from your computer via the microcontroller.
//
// In other words:
// Your computer <--> UART0 <--> MCU <--> UART1 <--> ESP8266 <--> INTERNET
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/espat"
	"tinygo.org/x/drivers/net"
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
	if connectToESP() {
		println("Connected to wifi adaptor.")
		adaptor.Echo(false)

		connectToAP()
	} else {
		println("")
		failMessage("Unable to connect to wifi adaptor.")
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
