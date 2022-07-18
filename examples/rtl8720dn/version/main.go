package main

import (
	"fmt"
	"time"

	"tinygo.org/x/drivers/examples/rtl8720dn"
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
	//rtl8720dn.Debug(true)
	rtl, err := rtl8720dn.Setup()
	if err != nil {
		return err
	}

	ver, err := rtl.Version()
	if err != nil {
		return nil
	}

	for {
		fmt.Printf("RTL8270DN Firmware Version: %s\r\n", ver)
		time.Sleep(10 * time.Second)
	}

	return nil
}
