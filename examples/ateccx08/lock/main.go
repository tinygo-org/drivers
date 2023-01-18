package main

import (
	"machine"

	"encoding/hex"
	"time"

	"tinygo.org/x/drivers/ateccx08"
)

func main() {
	time.Sleep(5 * time.Second)
	println("Looking for ATECCx08...")

	machine.I2C0.Configure(machine.I2CConfig{})

	atecc := ateccx08.New(machine.I2C0)
	atecc.Configure()

	if !atecc.Connected() {
		for {
			println("could not connect to ATECCx08")
			time.Sleep(time.Second)
		}
	}

	version, _ := atecc.Version()

	println(version.String(), "started")

	if !atecc.IsLocked() {
		for i := 10; i > 0; i-- {
			println(version.String(), "is not locked. Locking in", i, "seconds...")
			time.Sleep(time.Second)
		}

		// locks the Configuration zone... PERMANENTLY!
		atecc.Lock(0)
	}

	println(version.String(), "locked.")

	for {
		data, err := atecc.Random()
		if err != nil {
			println(err)
		}

		println(hex.EncodeToString(data[:]))
		time.Sleep(500 * time.Millisecond)
	}
}
