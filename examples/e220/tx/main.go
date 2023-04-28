// This is example of writing data to E220.
package main

import (
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
	addr := uint16(0xFFFF)
	ch := uint8(15)
	err = device.SetTxInfo(addr, ch, e220.TxMethodTransparent)
	if err != nil {
		fail(err.Error())
	}
	msg := "Hello,\nworld\n!!"
	for cnt := 0; ; cnt++ {
		_, err := fmt.Fprintf(device, "%s: %d\n", msg, cnt)
		if err != nil {
			fail(err.Error())
		}
		time.Sleep(time.Millisecond * 1000)
	}
}

func fail(msg string) {
	for {
		println(msg)
		time.Sleep(1 * time.Second)
	}
}
