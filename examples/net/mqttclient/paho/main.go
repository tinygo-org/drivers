// This example is an MQTT client built with the paho-mqtt package.  It sends
// machine.CPUFrequency() readings to the broker every second for 10 seconds.
//
// Note: It may be necessary to increase the stack size when using
// paho.mqtt.golang.  Use the -stack-size=4KB command line option.

//go:build ninafw || wioterminal || challenger_rp2040

package main

import (
	"fmt"
	"log"
	"machine"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var (
	ssid   string
	pass   string
	broker string = "tcp://test.mosquitto.org:1883"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Message %s received on topic %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection Lost: %s\n", err.Error())
}

func main() {
	waitSerial()

	link, _ := probe.Probe()

	err := link.NetConnect(&netlink.ConnectParams{
		Ssid:       ssid,
		Passphrase: pass,
	})
	if err != nil {
		log.Fatal(err)
	}

	clientId := "tinygo-client-" + randomString(10)
	fmt.Printf("ClientId: %s\n", clientId)

	options := mqtt.NewClientOptions()
	options.AddBroker(broker)
	options.SetClientID(clientId)
	options.SetDefaultPublishHandler(messagePubHandler)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectionLostHandler

	fmt.Printf("Connecting to MQTT broker at %s\n", broker)
	client := mqtt.NewClient(options)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	topic := "cpu/freq"
	token = client.Subscribe(topic, 1, nil)
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Printf("Subscribed to topic %s\n", topic)

	for i := 0; i < 10; i++ {
		freq := float32(machine.CPUFrequency()) / 1000000
		payload := fmt.Sprintf("%.02fMhz", freq)
		token = client.Publish(topic, 0, false, payload)
		if token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
		time.Sleep(time.Second)
	}

	client.Disconnect(100)

	for {
		select {}
	}
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

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}
