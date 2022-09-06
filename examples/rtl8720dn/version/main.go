package main

import (
	"machine"

	"fmt"
	"time"

	"tinygo.org/x/drivers/rtl8720dn"
)

var (
	debug = false
)

func main() {
	err := run()
	for err != nil {
		fmt.Printf("error: %s\r\n", err.Error())
		time.Sleep(5 * time.Second)
	}
}

func run() error {
	adaptor := rtl8720dn.New(machine.UART3, machine.PB24, machine.PC24, machine.RTL8720D_CHIP_PU)
	adaptor.Debug(debug)
	adaptor.Configure()

	ver, err := adaptor.Version()
	if err != nil {
		return nil
	}

	for {
		fmt.Printf("RTL8270DN Firmware Version: %s\r\n", ver)
		time.Sleep(10 * time.Second)
	}

	return nil
}
