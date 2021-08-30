// This example opens a TCP connection using a device with WiFiNINA firmware
// and sends a HTTP request to retrieve a webpage, based on the following
// Arduino example:
//
// https://github.com/arduino-libraries/WiFiNINA/blob/master/examples/WiFiWebClientRepeating/
//
// This example will not work with samd21 or other systems with less than 32KB
// of RAM.  Use the following if you want to run wifinina on samd21, etc.
//
// examples/wifinina/webclient
// examples/wifinina/tlsclient
//
package main

import (
	"bufio"
	"fmt"
	"machine"
	"strings"
	"time"

	"tinygo.org/x/drivers/net"
	"tinygo.org/x/drivers/net/http"
	"tinygo.org/x/drivers/wifinina"
)

// access point info
const ssid = ""
const pass = ""
const timeout = time.Minute

// Can specify a URL starting with http or https
const url = "https://raw.githubusercontent.com/tinygo-org/tinygo/release/LICENSE"

var (
	// these are the default pins for the Arduino Nano connected boards.
	// change these to connect to a different UART or pins for the ESP8266/ESP32
	spi = machine.NINA_SPI

	// this is the ESP chip that has the WIFININA firmware flashed on it
	device *wifinina.Device
)

var buf [0x1000]byte

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

	device = wifinina.New(spi,
		machine.NINA_CS,
		machine.NINA_ACK,
		machine.NINA_GPIO0,
		machine.NINA_RESETN)
	device.Configure().WithAccessPoint(ssid, pass, timeout)
}

func main() {

	if len(ssid) == 0 || len(pass) == 0 {
		for {
			println("Please set ssid and password for this example to work")
			time.Sleep(10 * time.Second)
		}
	}

	setup()
	http.SetBuf(buf[:])

	waitSerial()

	// You can send and receive cookies in the following way
	// 	import "tinygo.org/x/drivers/net/http/cookiejar"
	// 	jar, err := cookiejar.New(nil)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	client := &http.Client{Jar: jar}
	// 	http.DefaultClient = client

	cnt := 0
	for {
		// Various examples are as follows
		//
		// -- Get
		// 	resp, err := http.Get(url)
		//
		// -- Post
		// 	body := `cnt=12`
		// 	resp, err = http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(body))
		//
		// -- Post with JSON
		// 	body := `{"msg": "hello"}`
		// 	resp, err := http.Post(url, "application/json", strings.NewReader(body))

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("%s\r\n", err.Error())
			continue
		}

		fmt.Printf("%s %s\r\n", resp.Proto, resp.Status)
		for k, v := range resp.Header {
			fmt.Printf("%s: %s\r\n", k, strings.Join(v, " "))
		}
		fmt.Printf("\r\n")

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			fmt.Printf("%s\r\n", scanner.Text())
		}
		resp.Body.Close()

		cnt++
		fmt.Printf("-------- %d --------\r\n", cnt)
		time.Sleep(10 * time.Second)
	}

}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}
