package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/irremote"
)

var irCmdButtons = map[uint16]string{
	0x45: "POWER",
	0x47: "FUNC/STOP",
	0x46: "VOL+",
	0x44: "FAST BACK",
	0x40: "PAUSE",
	0x43: "FAST FORWARD",
	0x07: "DOWN",
	0x15: "VOL-",
	0x09: "UP",
	0x19: "EQ",
	0x0D: "ST/REPT",
	0x16: "0",
	0x0C: "1",
	0x18: "2",
	0x5E: "3",
	0x08: "4",
	0x1C: "5",
	0x5A: "6",
	0x42: "7",
	0x52: "8",
	0x4A: "9",
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
