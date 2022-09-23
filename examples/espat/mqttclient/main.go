// This is a sensor station that uses a ESP8266 or ESP32 running on the device UART1.
// It creates an MQTT connection that publishes a message every second
// to an MQTT broker.
//
// In other words:
// Your computer <--> UART0 <--> MCU <--> UART1 <--> ESP8266 <--> Internet <--> MQTT broker.
//
// You must install the Paho MQTT package to build this program:
//
//	go get -u github.com/eclipse/paho.mqtt.golang
package main

import (
	"machine"
	"math/rand"
	"time"

	"tinygo.org/x/drivers/espat"
	"tinygo.org/x/drivers/net/mqtt"
)

var (
	// access point info
	ssid string
	pass string
)

// IP address of the MQTT broker to use. Replace with your own info.
const server = "tcp://test.mosquitto.org:1883"

//const server = "ssl://test.mosquitto.org:8883"

// these are the default pins for the Arduino Nano33 IoT.
// change these to connect to a different UART or pins for the ESP8266/ESP32
var (
	uart = machine.UART2
	tx   = machine.PA22
	rx   = machine.PA23

	console = machine.Serial

	adaptor *espat.Device
	topic   = "tinygo"
)

func main() {
	time.Sleep(3000 * time.Millisecond)

	uart.Configure(machine.UARTConfig{TX: tx, RX: rx})
	rand.Seed(time.Now().UnixNano())

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

	opts := mqtt.NewClientOptions()
	opts.AddBroker(server).SetClientID("tinygo-client-" + randomString(10))

	println("Connecting to MQTT broker at", server)
	cl := mqtt.NewClient(opts)
	if token := cl.Connect(); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
	}

	for {
		println("Publishing MQTT message...")
		data := []byte("{\"e\":[{ \"n\":\"hello\", \"v\":101 }]}")
		token := cl.Publish(topic, 0, false, data)
		token.Wait()
		if token.Error() != nil {
			println(token.Error().Error())
		}

		time.Sleep(1000 * time.Millisecond)
	}

	// Right now this code is never reached. Need a way to trigger it...
	println("Disconnecting MQTT...")
	cl.Disconnect(100)

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

// Returns an int >= min, < max
func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

// Generate a random string of A-Z chars with len = l
func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(randomInt(65, 90))
	}
	return string(bytes)
}

func failMessage(msg string) {
	for {
		println(msg)
		time.Sleep(1 * time.Second)
	}
}
