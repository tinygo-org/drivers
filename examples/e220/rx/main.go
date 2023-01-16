// This is example of reading data from E220.
package main

import (
	"bufio"
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/e220"
)

var (
	uart = machine.DefaultUART
	m0   = machine.D4
	m1   = machine.D5
	tx   = machine.UART_TX_PIN
	rx   = machine.UART_RX_PIN
	aux  = machine.D6
)

func main() {
	device := e220.New(uart, m0, m1, tx, rx, aux, 9600)
	err := device.Configure(e220.Mode3)
	if err != nil {
		fail(err.Error())
	}

	response, err := device.ReadConfig()
	if err != nil {
		fail(err.Error())
	}
	println(fmt.Sprintf("read config: %X", response))

	device.SetMode(e220.Mode0)
	scanner := bufio.NewScanner(device)
	for scanner.Scan() {
		println(scanner.Text())
	}
	// The following are not reached until EOF is given or an error occurs
	if err := scanner.Err(); err != nil {
		fail(err.Error())
	}
}

func fail(msg string) {
	for {
		println(msg)
		time.Sleep(1 * time.Second)
	}
}
