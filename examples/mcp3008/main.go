// Connects to a MCP3008 ADC via SPI.
package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/mcp3008"
)

var (
	spi   = machine.SPI0
	csPin = machine.D12
)

func main() {
	spi.Configure(machine.SPIConfig{
		Frequency: 4000000,
		Mode:      3})

	adc := mcp3008.New(spi, csPin)
	adc.Configure()

	// get "CH0" aka "machine.ADC" interface to channel 0 from ADC.
	p := adc.CH0

	for {
		val := p.Get()
		println(val)
		time.Sleep(50 * time.Millisecond)
	}
}
