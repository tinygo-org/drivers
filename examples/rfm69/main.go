package main

// This example sends "123" to the remote node while receiving the same message in background

import (
	"machine"

	"fmt"
	"time"

	"tinygo.org/x/drivers/rfm69"
)

func rc(d *rfm69.Data) {
	fmt.Printf("Received from: %v, data %v\n", d.FromAddress, d.Data)
}
func main() {

	time.Sleep(time.Second * 5)
	fmt.Println("Starting ...")

	fmt.Println("Setting up SPI0")
	machine.SPI0.Configure(machine.SPIConfig{Mode: 0, LSBFirst: false, Frequency: 8000000})
	opt := rfm69.RFMOptions{
		NodeID:        10, // Change this to your node id
		NetworkID:     100,
		IsRfm69HCW:    true,
		ResetPin:      machine.NoPin,
		IrqPin:        machine.GP20,
		CsPin:         machine.GP17,
		OnReceive:     rc,
		EncryptionKey: "sampleEncryptKey",
	}
	fmt.Println("Setting up rfm69")
	radio, err := rfm69.NewDevice(machine.SPI0, &opt)
	if err != nil {
		println("Error ", err)
	}
	defer radio.Close()
	myData := &rfm69.Data{
		ToAddress:   1, // change this to remote node id
		FromAddress: opt.NodeID,
		Data:        []byte("123"),
		RequestAck:  true,
		SendAck:     false,
		Rssi:        -100,
	}

	fmt.Println("RFM is setup")
	for {
		radio.Send(myData)
		fmt.Println("Sent test packet ... ")
		time.Sleep(time.Second)
	}
}
