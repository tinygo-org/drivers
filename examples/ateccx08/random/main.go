package main

import (
	"machine"

	"crypto/rand"
	"encoding/hex"
	"time"

	"tinygo.org/x/drivers/ateccx08"
)

var atecc *ateccx08.Device

func main() {
	time.Sleep(5 * time.Second)
	println("Looking for ATECCx08...")

	machine.I2C0.Configure(machine.I2CConfig{})

	atecc = ateccx08.New(machine.I2C0)
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
		for {
			println(version.String(), "is not locked. Random numbers will not actually be random.")
			time.Sleep(time.Second)
		}
	}

	var result [13]byte
	for {
		rand.Read(result[:])
		encodedString := hex.EncodeToString(result[:])
		println(encodedString)
		time.Sleep(500 * time.Millisecond)
	}
}
