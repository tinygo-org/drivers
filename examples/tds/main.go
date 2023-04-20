package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/tds"
)

func main() {
	tds := tds.New(machine.A0, 5.0, 65535.0)
	machine.InitADC()
	tds.Configure()
	for {
		r, err := tds.GetTds(20.0)
		if err != nil {
			println("failed to read TDS, ", err.Error())
		}
		println("---------------------------------------------------------------")
		// cast as uint so that results aren't in scientific notation
		println("tds reading: ", uint(r), "ppm")
		println("---------------------------------------------------------------")
		println("ec reading:", uint(tds.GetElectricalConductance(20.0)))
		println("---------------------------------------------------------------")
		time.Sleep(time.Second * 5)
	}
}
