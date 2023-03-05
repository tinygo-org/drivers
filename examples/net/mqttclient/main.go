// This example is a MQTT client.  It sends machine.ReadTemparature() readings
// to the broker every second for 10 seconds.
//
// Note: It may be necessary to increase the stack size when using
// paho.mqtt.golang.  Use the -stack-size=4KB command line option.

package main

import (
	"fmt"
	"log"
	"machine"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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

	if err := netdev.NetConnect(); err != nil {
		log.Fatal(err)
	}

	options := mqtt.NewClientOptions()
	options.AddBroker(broker)
	options.SetClientID("tinygo_mqtt_example")
	options.SetDefaultPublishHandler(messagePubHandler)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectionLostHandler

	client := mqtt.NewClient(options)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	topic := "cpu/freq"
	token = client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic %s\n", topic)

	for i := 0; i < 10; i++ {
		freq := float32(machine.CPUFrequency()) / 1000000
		payload := fmt.Sprintf("%.02fMhz", freq)
		token = client.Publish(topic, 0, false, payload)
		token.Wait()
		time.Sleep(time.Second)
	}

	client.Disconnect(100)
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}
