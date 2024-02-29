package main

import (
	"time"

	"tinygo.org/x/drivers/espradio"
)

func main() {
	time.Sleep(time.Second * 1)

	println("initializing radio...")
	err := espradio.Enable(espradio.Config{
		Logging: espradio.LogLevelVerbose,
	})
	if err != nil {
		println("could not enable radio:", err)
	} else {
		println("enabled radio successfully")
	}
}
