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
	spicfg = machine.SPIConfig{
		Frequency: 250000,
		Mode:      0,
		SCK:       SPI_SCK_PIN,
		SDO:       SPI_TX_PIN,
		SDI:       SPI_RX_PIN,
	}
)

func main() {
	time.Sleep(time.Second)
	SPI_CS_PIN.Configure(machine.PinConfig{Mode: machine.PinOutput})
	err := spibus.Configure(spicfg)
	if err != nil {
		panic(err.Error())
	}
	sdcard := sd.NewSPICard(spibus, SPI_CS_PIN.Set)
	println("start init")
	err = sdcard.Init()
	if err != nil {
		panic("sd card init:" + err.Error())
	}
	// After initialization it's safe to increase SPI clock speed.
	csd := sdcard.CSD()
	kbps := csd.TransferSpeed().RateKilobits()
	spicfg.Frequency = uint32(kbps * 1000)
	err = spibus.Configure(spicfg)

	cid := sdcard.CID()
	fmt.Printf("name=%s\ncsd=\n%s\n", cid.ProductName(), csd.String())

	var buf [512]byte
	for i := 0; i < 11; i += 1 {
		err = sdcard.ReadBlocks(buf[:], 0)
		if err != nil {
			println("err reading block", i, ":", err.Error())
			continue
		}
		expectCRC := sd.CRC16(buf[:])
		fmt.Printf("block %d theircrc=%#x ourcrc=%#x:\n\t%#x\n", i, sdcard.LastReadCRC(), expectCRC, buf[:])
	}
}
