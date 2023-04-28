// This is example of passing through data.
package main

import (
	"io"
	"machine"
	"os"
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
	led  = machine.LED
)

func main() {
	device := e220.New(uart, m0, m1, tx, rx, aux, 9600)
	err := device.Configure(e220.Mode0)
	if err != nil {
		fail(err.Error())
	}
	addr := uint16(0xFFFF)
	ch := uint8(15)
	err = device.SetTxInfo(addr, ch, e220.TxMethodTransparent)
	if err != nil {
		fail(err.Error())
	}
	go func() {
		io.Copy(os.Stdout, device)
	}()
	w := io.MultiWriter(os.Stdout, device)
	io.Copy(w, os.Stdin)
}
func fail(msg string) {
	for {
		println(msg)
		time.Sleep(1 * time.Second)
	}
}
