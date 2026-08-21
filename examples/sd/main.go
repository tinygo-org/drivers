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

	bd, err := sd.NewBlockDevice(sdcard, csd.ReadBlockLen(), csd.NumberOfBlocks())
	if err != nil {
		panic("block device creation:" + err.Error())
	}
	var mc MemChecker

	ok, badBlkIdx, err := mc.MemCheck(bd, 2, 100)
	if err != nil {
		panic("memcheck:" + err.Error())
	}
	if !ok {
		println("bad block", badBlkIdx)
	} else {
		println("memcheck ok")
	}
}

type MemChecker struct {
	rdBuf    []byte
	storeBuf []byte
	wrBuf    []byte
}

func (mc *MemChecker) MemCheck(bd *sd.BlockDevice, blockIdx, numBlocks int64) (memOK bool, badBlockIdx int64, err error) {
	size := bd.BlockSize() * numBlocks
	if len(mc.rdBuf) < int(size) {
		mc.rdBuf = make([]byte, size)
		mc.wrBuf = make([]byte, size)
		mc.storeBuf = make([]byte, size)
		for i := range mc.wrBuf {
			mc.wrBuf[i] = byte(i)
		}
	}
	// Start by storing the original block contents.
	_, err = bd.ReadAt(mc.storeBuf, blockIdx)
	if err != nil {
		return false, blockIdx, err
	}

	// Write the test pattern.
	_, err = bd.WriteAt(mc.wrBuf, blockIdx)
	if err != nil {
		return false, blockIdx, err
	}
	// Read back the test pattern.
	_, err = bd.ReadAt(mc.rdBuf, blockIdx)
	if err != nil {
		return false, blockIdx, err
	}
	for j := 0; j < len(mc.rdBuf); j++ {
		// Compare the read back data with the test pattern.
		if mc.rdBuf[j] != mc.wrBuf[j] {
			badBlock := blockIdx + int64(j)/bd.BlockSize()
			return false, badBlock, nil
		}
		mc.rdBuf[j] = 0
	}
	// Leave the card in it's previous state.
	_, err = bd.WriteAt(mc.storeBuf, blockIdx)
	return true, -1, nil
}
