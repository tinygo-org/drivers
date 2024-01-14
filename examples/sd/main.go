package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/sd"
)

const (
	SPI_RX_PIN  = machine.GP16
	SPI_TX_PIN  = machine.GP19
	SPI_SCK_PIN = machine.GP18
	SPI_CS_PIN  = machine.GP15
)

var (
	spibus = machine.SPI0
)

func main() {
	time.Sleep(time.Second)
	SPI_CS_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	err := spibus.Configure(machine.SPIConfig{
		Frequency: 250000,
		Mode:      0,
		SCK:       SPI_SCK_PIN,
		SDO:       SPI_TX_PIN,
		SDI:       SPI_RX_PIN,
	})
	if err != nil {
		panic(err.Error())
	}
	sdcard := sd.NewCard(spibus, SPI_CS_PIN.Set)

	err = sdcard.Init()
	if err != nil {
		panic(err.Error())
	}
	csd := sdcard.CSD()
	fmt.Printf("cid=%+v\ncsd=%+v\ncsdfmt=\n%s\ndone!", sdcard.CID(), csd, csd.String())

	var buf [512]byte
	for i := 1; i < 11; i += 1 {
		time.Sleep(time.Millisecond)
		err = sdcard.ReadBlock(uint32(i), buf[:])
		if err != nil {
			println("err reading block", i, ":", err.Error())
			continue
		}
		fmt.Printf("block %d crc=%#x:\n\t%#x\n", i, sdcard.LastReadCRC(), buf[:])
	}
}
