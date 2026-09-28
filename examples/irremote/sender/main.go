package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/irremote"
)

var (
	pinIROut = machine.GPIO17
	pwmIROut = machine.PWM0
	ir       irremote.SenderDevice
)

const (
	irAddress = 0x00
	irCmdPwr  = 0x45
)

func main() {
	ir = irremote.NewSender(irremote.SenderConfig{
		Pin:       pinIROut,
		PWM:       pwmIROut,
		DutyCycle: 33,
	})
	if err := ir.Configure(); err != nil {
		println(err.Error())
		return
	}
	for {
		// Send one frame, then hold the button down for half a second.
		ir.SendNEC(irAddress, irCmdPwr, true)
		time.Sleep(time.Millisecond * 500)
		ir.StopNECRepeats()
		time.Sleep(time.Second * 2)
	}
}
