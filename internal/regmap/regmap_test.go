package regmap

import (
	"fmt"

	"tinygo.org/x/drivers"
)

func ExampleDevice8Txer() {
	// Initialization.
	var dtx Device8Txer
	dtx.SetTxBuffers(make([]byte, 256), make([]byte, 256))

	// Usage.
	const (
		defaultAddr = 65
		REG_WRITE   = 0x1f
		IOCTL_CALL  = 0xc0
	)
	tx, err := dtx.Tx(REG_WRITE)
	if err != nil {
		panic(err)
	}
	err = tx.AddWriteData(IOCTL_CALL, 0x80, 0x80)
	if err != nil {
		panic(err)
	}
	var bus drivers.I2C
	readData, err := tx.DoTxI2C(bus, defaultAddr, 20)
	if err != nil {
		panic(err)
	}
	fmt.Println(readData)
}
