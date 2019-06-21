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
)

func main() {
	uart.Configure(machine.UARTConfig{TX: tx, RX: rx})

	time.Sleep(3000 * time.Millisecond)

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

	// now make TCP connection
	raddr, _ := adaptor.ResolveTCPAddr("tcp", server)
	if raddr != nil {
		println("The IP address from DNS lookup is:")
		println(string(raddr.IP))
		println(raddr.Port)
	}
	laddr := &espat.TCPAddr{Port: 1883}

	console.Write([]byte("Dialing TCP connection...\r\n"))
	conn, err = adaptor.DialTCP("tcp", laddr, raddr)
	if err != nil {
		println("tcp connect error")
		println(err.Error())
	}

	err = connectToMQTTServer()

	for {
		console.Write([]byte("Publishing MQTT packets...\r\n"))

		publish := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
		publish.Qos = 0
		publish.TopicName = "tinygo"
		publish.Payload = []byte("Hello, mqtt\r\n")
		publish.MessageID = mid
		mid++

		publish.Write(conn)

		time.Sleep(1000 * time.Millisecond)
	}

	// Right now this code is never reached. Need a way to trigger it...
	console.Write([]byte("Disconnecting TCP...\r\n"))
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

func connectToMQTTServer() error {
	// send the MQTT connect message
	connectPkt := packets.NewControlPacket(packets.Connect).(*packets.ConnectPacket)
	connectPkt.Qos = 0
	// connectPkt.Username = "tinygo"
	// connectPkt.Password = []byte("1234")
	connectPkt.ClientIdentifier = "tinygo-client"
	connectPkt.ProtocolVersion = 4
	connectPkt.ProtocolName = "MQTT"
	connectPkt.Keepalive = 30

	console.Write([]byte("Sending MQTT connect...\r\n"))
	err := connectPkt.Write(conn)
	if err != nil {
		println("mqtt connect error")
		println(err.Error())
		return err
	}

	console.Write([]byte("Waiting for MQTT connect...\r\n"))
	// TODO: handle timeout
	for {
		packet, _ := packets.ReadPacket(conn)

		if packet != nil {
			_, ok := packet.(*packets.ConnackPacket)
			if ok {
				console.Write([]byte("Connected to MQTT server.\r\n"))
				println(packet.String())
				return nil
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}
