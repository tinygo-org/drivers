package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/pca9685"
)

func main() {
	const (
		// Default address on most breakout boards.
		pcaAddr = 0x40
	)
	err := machine.I2C0.Configure(machine.I2CConfig{})
	if err != nil {
		panic(err.Error())
	}
	d := pca9685.New(machine.I2C0, 0x40)
	err = d.IsConnected()
	if err != nil {
		panic(err.Error())
	}
	err = d.Configure(pca9685.PWMConfig{Period: 1e9 / 200}) // 200Hz PWM
	if err != nil {
		panic(err.Error())
	}

	var value uint32
	step := d.Top() / 5
	for {
		for value = 0; value <= d.Top(); value += step {
			d.SetAll(value)
			dc := 100 * value / d.Top()
			println("set dc @", dc, "%")
			time.Sleep(800 * time.Millisecond)
		}
	}
}

// ScanI2CDev finds I2C devices on the bus and rreturns them inside
// a slice. If slice is nil then no devices were found.
func ScanI2CDev(bus machine.I2C) (addrs []uint8) {
	var addr, count uint8
	var err error
	w := []byte{1}
	// Count devices in first scan
	for addr = 1; addr < 127; addr++ {
		err = bus.Tx(uint16(addr), w, nil)
		if err == nil {
			count++
		}
	}
	if count == 0 {
		return nil
	}
	// Allocate slice and populate slice with addresses
	addrs = make([]uint8, count)
	count = 0
	for addr = 1; addr < 127; addr++ {
		err = bus.Tx(uint16(addr), w, nil)
		if err == nil && count < uint8(len(addrs)) {
			addrs[count] = addr
			count++
		}
	}
	return addrs
}
