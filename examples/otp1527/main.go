package main

import (
	"fmt"
	"machine"

	"tinygo.org/x/drivers/otp1527"
)

func main() {
	d := otp1527.NewDecoder(machine.Pin(3), -1)
	for {
		v := <-d.Out()
		println(fmt.Sprintf("RECV: 0x%06x", v.Data))
	}
}
