package mcp9808

import (
	"fmt"
	"time"
)

func main() {
	r := raspi.NewAdaptor()
	bus := i2c.NewBus(r)
	sensor, err := NewMCP9808(bus, _MCP9808_DEFAULT_ADDRESS)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		temp, err := sensor.Temperature()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Temperature: %.2f°C\n", temp)

		time.Sleep(1 * time.Second)
	}
}
