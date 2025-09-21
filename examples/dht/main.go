//go:build baremetal && tinygo

package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/dht"
	"tinygo.org/x/drivers/tinygo"
)

func main() {
	pin := tinygo.New(machine.D6)
	dhtSensor := dht.New(pin, dht.DHT11)
	for {
		temp, hum, err := dhtSensor.Measurements()
		if err == nil {
			fmt.Printf("Temperature: %02d.%d°C, Humidity: %02d.%d%%\n", temp/10, temp%10, hum/10, hum%10)
		} else {
			fmt.Printf("Could not take measurements from the sensor: %s\n", err.Error())
		}
		// Measurements cannot be updated only 2 seconds. More frequent calls will return the same value
		time.Sleep(time.Second * 2)
	}
}
