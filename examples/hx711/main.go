package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/hx711"
)

const (
	clockOutPin     = machine.D3
	dataInPin       = machine.D2
	gainAndChannel  = hx711.A128           // only the first channel A is used
	tickSleep       = 1 * time.Microsecond // set it to zero for slow MCU's
	calibrationWait = 10 * time.Second
	cycleTime       = 1 * time.Second
)

// please adjust to your load used for calibration
const (
	setLoad = 100 // used unit will equal the measured unit
	unit    = "gram"
)

func main() {
	time.Sleep(5 * time.Second) // wait for monitor connection

	cfg := hx711.DefaultConfig
	cfg.TickSleep = tickSleep

	clockOutPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	dataInPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	clockOutPinSetState := func(v bool) error { clockOutPin.Set(v); return nil }
	dataInPinState := func() (bool, error) { return dataInPin.Get(), nil }

	sensor := hx711.New(clockOutPinSetState, dataInPinState, gainAndChannel)
	if err := sensor.Configure(&cfg); err != nil {
		println("Configure failed")
		panic(err)
	}

	println("Please remove the mass completely for zeroing within", calibrationWait.String())
	time.Sleep(calibrationWait)
	println("Zero starts")
	if err := sensor.Zero(false); err != nil {
		println("Zeroing failed")
		panic(err)
	}

	println("Please apply the load (", setLoad, unit+" ) for calibration within", calibrationWait.String())
	time.Sleep(calibrationWait)
	println("Calibration starts")
	if err := sensor.Calibrate(setLoad, false); err != nil {
		println("Calibration failed")
		panic(err)
	}

	offs, factor := sensor.OffsetAndCalibrationFactor(false)
	println("Calibration done completely, offset:", offs, "factor:", factor)

	println("Measurement starts")
	for {
		if err := sensor.Update(0); err != nil {
			println("Sensor update failed", err.Error())
		}

		v1, _ := sensor.Values()

		println("Mass:", v1, unit)

		time.Sleep(cycleTime)
	}
}
