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
	sdcard := sd.NewSPICard(spibus, SPI_CS_PIN.Set)

	err = sdcard.Init()
	if err != nil {
		panic(err.Error())
	}
	cid := sdcard.CID()
	pname := cid.ProductName()
	csd := sdcard.CSD()

	valid := csd.IsValid()
	if !valid {
		data := csd.RawCopy()
		crc := sd.CRC7(data[:15])
		always1 := data[15]&(1<<7) != 0
		fmt.Printf("ourCRC7=%#b theirCRC7=%#b for data %d\n", crc, csd.CRC7(), data[:15])
		println("CSD not valid got", crc, "want", csd.CRC7(), "always1:", always1)
		return
	} else {
		println("CSD valid!")
	}

	fmt.Printf("name=%s\ncsd=\n%s\n", pname, csd.String())

	var buf [512]byte
	for i := 0; i < 11; i += 1 {
		time.Sleep(100 * time.Millisecond)
		err = sdcard.ReadBlock(int64(i), buf[:])
		if err != nil {
			println("err reading block", i, ":", err.Error())
			continue
		}
		expectCRC := sd.CRC16(buf[:])
		fmt.Printf("block %d theircrc=%#x ourcrc=%#x:\n\t%#x\n", i, sdcard.LastReadCRC(), expectCRC, buf[:])
	}
}
