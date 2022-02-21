package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/irremote"
)

var irCmdButtons = map[uint16]string{
	0xA2: "POWER",
	0xE2: "FUNC/STOP",
	0x62: "VOL+",
	0x22: "FAST BACK",
	0x02: "PAUSE",
	0xC2: "FAST FORWARD",
	0xE0: "DOWN",
	0xA8: "VOL-",
	0x90: "UP",
	0x98: "EQ",
	0xB0: "ST/REPT",
	0x68: "0",
	0x30: "1",
	0x18: "2",
	0x7A: "3",
	0x10: "4",
	0x38: "5",
	0x5A: "6",
	0x42: "7",
	0x4A: "8",
	0x52: "9",
}

var (
	pinIRIn = machine.GP26
	ir      irremote.ReceiverDevice
)

func setupPins() {
	ir = irremote.NewReceiver(pinIRIn)
	ir.Configure()
}

func irCallback(data irremote.Data) {
	msg := "Command: " + irCmdButtons[data.Command]
	if data.Flags&irremote.DataFlagIsRepeat != 0 {
		msg = msg + " (REPEAT)"
	}
	println(msg)
}

func main() {
	setupPins()
	ir.SetCommandHandler(irCallback)
	for {
		time.Sleep(time.Millisecond * 10)
	}
}
