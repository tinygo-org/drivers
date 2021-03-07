package main

import (
	"machine"
	"time"

	"github.com/Nerzal/drivers/wifinina"
)

// access point info
const ssid = "NoobyGames"
const pass = "IchHasseLangeWlanZugangsDaten1312!"

// these are the default pins for the Arduino Nano33 IoT.
// change these to connect to a different UART or pins for the ESP8266/ESP32
var (

	// these are the default pins for the Arduino Nano33 IoT.
	uart = machine.UART2
	tx   = machine.NINA_TX
	rx   = machine.NINA_RX
	spi  = machine.NINA_SPI

	// this is the ESP chip that has the WIFININA firmware flashed on it
	adaptor = &wifinina.Device{
		SPI:   spi,
		CS:    machine.NINA_CS,
		ACK:   machine.NINA_ACK,
		GPIO0: machine.NINA_GPIO0,
		RESET: machine.NINA_RESETN,
	}

	console = machine.UART0
	topic   = "tinygo"
)

func main() {
	// Configure SPI for 8Mhz, Mode 0, MSB First
	spi.Configure(machine.SPIConfig{
		Frequency: 8 * 1e6,
		SDO:       machine.NINA_SDO,
		SDI:       machine.NINA_SDI,
		SCK:       machine.NINA_SCK,
	})

	// Init esp8266/esp32
	adaptor.Configure()
	time.Sleep(3 * time.Second)

	connectToAP()

	for {
		println("pinging")
		resp, err := adaptor.Ping(wifinina.IPAddress("8.8.8.8"), 250)
		if err != nil {
			println("failed to ping:", err)
			time.Sleep(time.Second)
			continue
		}

		println("ping response after :", resp, "ms")

		time.Sleep(time.Second)
	}
}

// connect to access point
func connectToAP() {
	time.Sleep(2 * time.Second)
	println("Connecting to " + ssid)
	adaptor.SetPassphrase(ssid, pass)
	for st, _ := adaptor.GetConnectionStatus(); st != wifinina.StatusConnected; {
		println("Connection status: " + st.String())
		time.Sleep(1 * time.Second)
		st, _ = adaptor.GetConnectionStatus()
	}
	println("Connected.")
	time.Sleep(2 * time.Second)
	ip, _, _, err := adaptor.GetIP()
	for ; err != nil; ip, _, _, err = adaptor.GetIP() {
		println(err.Error())
		time.Sleep(1 * time.Second)
	}
	println(ip.String())
}
