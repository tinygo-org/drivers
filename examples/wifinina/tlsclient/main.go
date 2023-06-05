// This example opens a TCP connection using a device with WiFiNINA firmware
// and sends a HTTPS request to retrieve a webpage
//
// You shall see "strict-transport-security" header in the response,
// this confirms communication is indeed over HTTPS
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security
package main

import (
	"fmt"
	"machine"
	"strings"
	"time"

	"tinygo.org/x/drivers/net"
	"tinygo.org/x/drivers/net/tls"
	"tinygo.org/x/drivers/wifinina"
)

var (
	// access point info
	ssid string
	pass string
)

// IP address of the server aka "hub". Replace with your own info.
const server = "tinygo.org"

// these are the default pins for the Arduino Nano33 IoT.
// change these to connect to a different UART or pins for the ESP8266/ESP32
var (

	// these are the default pins for the Arduino Nano33 IoT.
	spi = machine.NINA_SPI

	// this is the ESP chip that has the WIFININA firmware flashed on it
	adaptor *wifinina.Device
)

var buf [256]byte

var lastRequestTime time.Time
var conn net.Conn

func setup() {
	// Configure SPI for 8Mhz, Mode 0, MSB First
	spi.Configure(machine.SPIConfig{
		Frequency: 8 * 1e6,
		SDO:       machine.NINA_SDO,
		SDI:       machine.NINA_SDI,
		SCK:       machine.NINA_SCK,
	})

	adaptor = wifinina.New(spi,
		machine.NINA_CS,
		machine.NINA_ACK,
		machine.NINA_GPIO0,
		machine.NINA_RESETN)
	adaptor.Configure()
}

func main() {

	setup()

	waitSerial()

	connectToAP()
	displayIP()

	for {
		readConnection()
		if time.Now().Sub(lastRequestTime).Milliseconds() >= 10000 {
			makeHTTPSRequest()
		}
	}

}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}

func readConnection() {
	if conn != nil {
		for n, err := conn.Read(buf[:]); n > 0; n, err = conn.Read(buf[:]) {
			if err != nil {
				println("Read error: " + err.Error())
			} else {
				print(string(buf[0:n]))
			}
		}
	}
}

func makeHTTPSRequest() {

	var err error
	if conn != nil {
		conn.Close()
	}

	message("\r\n---------------\r\nDialing TCP connection")
	conn, err = tls.Dial("tcp", server, nil)
	for ; err != nil; conn, err = tls.Dial("tcp", server, nil) {
		message("Connection failed: " + err.Error())
		time.Sleep(5 * time.Second)
	}
	println("Connected!\r")

	print("Sending HTTPS request...")
	fmt.Fprintln(conn, "GET / HTTP/1.1")
	fmt.Fprintln(conn, "Host:", strings.Split(server, ":")[0])
	fmt.Fprintln(conn, "User-Agent: TinyGo")
	fmt.Fprintln(conn, "Connection: close")
	fmt.Fprintln(conn)
	println("Sent!\r\n\r")

	lastRequestTime = time.Now()
}

const retriesBeforeFailure = 3

// connect to access point
func connectToAP() {
	time.Sleep(2 * time.Second)
	var err error
	for i := 0; i < retriesBeforeFailure; i++ {
		println("Connecting to " + ssid)
		err = adaptor.ConnectToAccessPoint(ssid, pass, 10*time.Second)
		if err == nil {
			println("Connected.")

			return
		}
	}

	// error connecting to AP
	failMessage(err.Error())
}

func displayIP() {
	ip, _, _, err := adaptor.GetIP()
	for ; err != nil; ip, _, _, err = adaptor.GetIP() {
		message(err.Error())
		time.Sleep(1 * time.Second)
	}
	message("IP address: " + ip.String())
}

func message(msg string) {
	println(msg, "\r")
}

func failMessage(msg string) {
	for {
		println(msg)
		time.Sleep(1 * time.Second)
	}
}
