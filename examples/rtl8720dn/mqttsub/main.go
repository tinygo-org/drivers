// This is a sensor station that uses a RTL8720DN running on the device UART2.
// It creates an MQTT connection that publishes a message every second
// to an MQTT broker.
//
// In other words:
// Your computer <--> USB-CDC <--> MCU <--> UART2 <--> RTL8720DN <--> Internet <--> MQTT broker.
//
// You must also install the Paho MQTT package to build this program:
//
//	go get -u github.com/eclipse/paho.mqtt.golang
//
// You can check that mqttpub/mqttsub is running successfully with the following command.
//
//	mosquitto_sub -h test.mosquitto.org -t tinygo/tx
//	mosquitto_pub -h test.mosquitto.org -t tinygo/rx -m "hello world"
package main

import (
	"machine"

	"fmt"
	"math/rand"
	"time"

	"tinygo.org/x/drivers/net"
	"tinygo.org/x/drivers/net/mqtt"
	"tinygo.org/x/drivers/rtl8720dn"
)

// You can override the setting with the init() in another source code.
// func init() {
//    ssid = "your-ssid"
//    pass = "your-password"
//    debug = true
//    server = "tinygo.org"
// }

var (
	ssid   string
	pass   string
	server string = "tcp://test.mosquitto.org:1883"
	debug         = false
)

var buf [0x400]byte

var lastRequestTime time.Time
var conn net.Conn
var adaptor *rtl8720dn.Driver

func main() {
	err := run()
	for err != nil {
		fmt.Printf("error: %s\r\n", err.Error())
		time.Sleep(5 * time.Second)
	}
}

// change these to connect to a different UART or pins for the ESP8266/ESP32
var (
	cl      mqtt.Client
	topicTx = "tinygo/tx"
	topicRx = "tinygo/rx"
)

func subHandler(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("[%s]  ", msg.Topic())
	fmt.Printf("%s\r\n", msg.Payload())
}

func run() error {
	// change the UART and pins as needed for platforms other than the WioTerminal.
	adaptor = rtl8720dn.New(machine.UART3, machine.PB24, machine.PC24, machine.RTL8720D_CHIP_PU)
	adaptor.Debug(debug)
	adaptor.Configure()

	err := adaptor.ConnectToAccessPoint(ssid, pass, 10*time.Second)
	if err != nil {
		return err
	}

	ip, subnet, gateway, err := adaptor.GetIP()
	if err != nil {
		return err
	}
	fmt.Printf("IP Address : %s\r\n", ip)
	fmt.Printf("Mask       : %s\r\n", subnet)
	fmt.Printf("Gateway    : %s\r\n", gateway)

	rand.Seed(time.Now().UnixNano())

	opts := mqtt.NewClientOptions()
	opts.AddBroker(server).SetClientID("tinygo-client-" + randomString(10))

	println("Connecting to MQTT broker at", server)
	cl = mqtt.NewClient(opts)
	if token := cl.Connect(); token.Wait() && token.Error() != nil {
		failMessage(token.Error().Error())
	}

	// subscribe
	token := cl.Subscribe(topicRx, 0, subHandler)
	token.Wait()
	if token.Error() != nil {
		failMessage(token.Error().Error())
	}

	go publishing()

	select {}

	// Right now this code is never reached. Need a way to trigger it...
	println("Disconnecting MQTT...")
	cl.Disconnect(100)

	println("Done.")

	return nil
}

func publishing() {
	for i := 0; ; i++ {
		println("Publishing MQTT message...")
		data := []byte(fmt.Sprintf(`{"e":[{"n":"hello %d","v":101}]}`, i))
		token := cl.Publish(topicTx, 0, false, data)
		token.Wait()
		if token.Error() != nil {
			println(token.Error().Error())
		}

		time.Sleep(20 * 100 * time.Millisecond)
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

func failMessage(msg string) {
	for {
		println(msg)
		time.Sleep(1 * time.Second)
	}
}
