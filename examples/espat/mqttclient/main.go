// This is a sensor station that uses a ESP8266 or ESP32 running on the device UART1.
// It creates an MQTT connection that publishes a message every second
// to an MQTT broker.
//
// In other words:
// Your computer <--> UART0 <--> MCU <--> UART1 <--> ESP8266 <--> Internet <--> MQTT broker.
//
// You must install the Paho MQTT package to build this program:
//
// 		go get github.com/eclipse/paho.mqtt.golang
//
package main

import (
	"machine"
	"math/rand"
	"time"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"tinygo.org/x/drivers/espat"
)

// access point info
const ssid = "YOURSSID"
const pass = "YOURPASS"

// IP address of the MQTT broker to use. Replace with your own info.
const server = "test.mosquitto.org:1883"

// change these to connect to a different UART or pins for the ESP8266/ESP32
var (
	uart = machine.UART1
	tx   = machine.D10
	rx   = machine.D11

	console = machine.UART0

	adaptor *espat.Device
	conn    *espat.TCPSerialConn
	err     error
	mid     uint16
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
		println("Unable to connect to wifi adaptor.")
		return
	}

	// now make TCP connection
	raddr, _ := adaptor.ResolveTCPAddr("tcp", server)
	if raddr != nil {
		println("The IP address from DNS lookup is:")
		println(string(raddr.IP))
		println(raddr.Port)
	}
	laddr := &espat.TCPAddr{Port: 1883}

	println("Dialing TCP connection...")
	conn, err = adaptor.DialTCP("tcp", laddr, raddr)
	if err != nil {
		println("tcp connect error")
		println(err.Error())
	}

	err = connectToMQTTServer()

	for {
		publishToMQTT()

		time.Sleep(1000 * time.Millisecond)
	}

	// Right now this code is never reached. Need a way to trigger it...
	println("Disconnecting TCP...")
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
	println("Connecting to wifi network...")

	adaptor.SetWifiMode(espat.WifiModeClient)
	adaptor.ConnectToAP(ssid, pass, 10)

	println("Connected.")
	println(adaptor.GetClientIP())
}

func connectToMQTTServer() error {
	// send the MQTT connect message
	connectPkt := packets.NewControlPacket(packets.Connect).(*packets.ConnectPacket)
	connectPkt.Qos = 0
	// connectPkt.Username = "tinygo"
	// connectPkt.Password = []byte("1234")
	connectPkt.ClientIdentifier = "tinygo-client-" + randomString(10)
	connectPkt.ProtocolVersion = 4
	connectPkt.ProtocolName = "MQTT"
	connectPkt.Keepalive = 30

	println("Sending MQTT connect...")
	err := connectPkt.Write(conn)
	if err != nil {
		println("mqtt connect error")
		println(err.Error())
		return err
	}

	println("Waiting for MQTT connect...")
	// TODO: handle timeout
	for {
		packet, _ := packets.ReadPacket(conn)

		if packet != nil {
			_, ok := packet.(*packets.ConnackPacket)
			if ok {
				println("Connected to MQTT server.")
				println(packet.String())
				return nil
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func publishToMQTT() error {
	println("Publishing MQTT message...")

	publish := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
	publish.Qos = 0
	publish.TopicName = topic
	publish.Payload = []byte("Hello, mqtt\r\n")
	publish.MessageID = mid
	mid++

	return publish.Write(conn)
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
