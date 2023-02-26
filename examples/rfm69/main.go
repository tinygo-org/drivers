package main

// This example demonstrates some features of the PWM support.

import (
	"machine"

	"fmt"
	"time"

	"tinygo.org/x/drivers/rfm69"
)

func main() {

	time.Sleep(time.Second * 5)
	fmt.Println("Starting ...")

	fmt.Println("Setting up SPI0")
	machine.SPI0.Configure(machine.SPIConfig{Mode: 0, LSBFirst: false})
	opt := rfm69.RFMOptions{
		NodeID:     1,
		NetworkID:  1,
		IsRfm69HCW: true,
		ResetPin:   machine.NoPin,
		IrqPin:     machine.GP20,
		CsPin:      machine.GP17,
	}
	fmt.Println("Setting up rfm69")
	radio, err := rfm69.NewDevice(machine.SPI0, &opt)
	if err != nil {
		println("Error ", err)
	}
	defer radio.Close()
	myData := &rfm69.Data{
		ToAddress:   10,
		FromAddress: 1,
		Data:        []byte("123"),
		RequestAck:  false,
		SendAck:     false,
		Rssi:        100,
	}

	fmt.Println("RFM is setup")
	radio.Send(myData)
	fmt.Println("Sent test packet ... ")
}
