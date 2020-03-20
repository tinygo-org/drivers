package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/rcswitch"
)

const (
	rcPin = machine.D8
)

func main() {
	rc := rcswitch.New(rcPin)
	rc.Configure(rcswitch.Config{Group: "11011", Device: "00100"})
	rc.On()
	time.Sleep(1 * time.Second)
	rc.Off()
}
