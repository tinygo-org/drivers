// This example is an MQTT client built with the natiu-mqtt package.  It sends
// machine.CPUFrequency() readings to the broker every second for 10 seconds.
//
// Note: It may be necessary to increase the stack size when using
// paho.mqtt.golang.  Use the -stack-size=4KB command line option.

//go:build ninafw || wioterminal || challenger_rp2040

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"machine"
	"math/rand"
	"net"
	"time"

	mqtt "github.com/soypat/natiu-mqtt"
	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var (
	ssid   string
	pass   string
	broker string = "test.mosquitto.org:1883"
	topic  string = "cpu/freq"
)

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

	// Get a transport for MQTT packets
	fmt.Printf("Connecting to MQTT broker at %s\n", broker)
	conn, err := net.Dial("tcp", broker)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Create new client
	client := mqtt.NewClient(mqtt.ClientConfig{
		Decoder: mqtt.DecoderNoAlloc{make([]byte, 1500)},
		OnPub: func(_ mqtt.Header, _ mqtt.VariablesPublish, r io.Reader) error {
			message, _ := io.ReadAll(r)
			fmt.Printf("Message %s received on topic %s\n", string(message), topic)
			return nil
		},
	})

	// Connect client
	var varconn mqtt.VariablesConnect
	varconn.SetDefaultMQTT([]byte(clientId))
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx, conn, &varconn)
	if err != nil {
		log.Fatal("failed to connect: ", err)
	}

	// Subscribe to topic
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Subscribe(ctx, mqtt.VariablesSubscribe{
		PacketIdentifier: 23,
		TopicFilters: []mqtt.SubscribeRequest{
			{TopicFilter: []byte(topic), QoS: mqtt.QoS0},
		},
	})
	if err != nil {
		log.Fatal("failed to subscribe to", topic, err)
	}
	fmt.Printf("Subscribed to topic %s\n", topic)

	// Publish on topic
	pubFlags, _ := mqtt.NewPublishFlags(mqtt.QoS0, false, false)
	pubVar := mqtt.VariablesPublish{
		TopicName: []byte(topic),
	}

	for i := 0; i < 10; i++ {
		if !client.IsConnected() {
			log.Fatal("client disconnected: ", client.Err())
		}

		freq := float32(machine.CPUFrequency()) / 1000000
		payload := fmt.Sprintf("%.02fMhz", freq)

		pubVar.PacketIdentifier++
		err = client.PublishPayload(pubFlags, pubVar, []byte(payload))
		if err != nil {
			log.Fatal("error transmitting message: ", err)
		}

		time.Sleep(time.Second)

		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		err = client.HandleNext()
		if err != nil {
			log.Fatal("handle next: ", err)
		}

	}

	client.Disconnect(errors.New("disconnected gracefully"))

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
