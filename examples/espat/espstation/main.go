// This is a sensor station that uses a ESP8266 or ESP32 running on the device UART1.
// It creates a UDP connection you can use to get info to/from your computer via the microcontroller.
//
// In other words:
// Your computer <--> UART0 <--> MCU <--> UART1 <--> ESP8266
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/espat"
	"tinygo.org/x/drivers/net"
)

var (
	// access point info
	ssid string
	pass string
)

// IP address of the listener aka "hub". Replace with your own info.
const hubIP = "0.0.0.0"

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

	// Init esp8266/esp32
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

	// now make UDP connection
	ip := net.ParseIP(hubIP)
	raddr := &net.UDPAddr{IP: ip, Port: 2222}
	laddr := &net.UDPAddr{Port: 2222}

	println("Dialing UDP connection...")
	conn, _ := net.DialUDP("udp", laddr, raddr)

	for {
		// send data
		println("Sending data...")
		conn.Write([]byte("hello\r\n"))

		time.Sleep(1000 * time.Millisecond)
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

func failMessage(msg string) {
	for {
		println(msg)
		time.Sleep(1 * time.Second)
	}
}
