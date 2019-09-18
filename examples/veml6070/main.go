package main

import (
	"time"

	"machine"

	"tinygo.org/x/drivers/veml6070"
)

func main() {
	machine.I2C0.Configure(machine.I2CConfig{})
	sensor := veml6070.New(machine.I2C0)

	if !sensor.Configure() {
		println("VEML6070 could not be configured")
		return
	}

	println("VEML6070 configured")

	for {
		intensity, _ := sensor.ReadUVALightIntensity()
		println("UVA light intensity:", float32(intensity)/1000.0, "W/(m*m)")

		switch sensor.GetEstimatedRiskLevel(intensity) {
		case veml6070.UVI_RISK_LOW:
			println("UV risk level: low")
		case veml6070.UVI_RISK_MODERATE:
			println("UV risk level: moderate")
		case veml6070.UVI_RISK_HIGH:
			println("UV risk level: high")
		case veml6070.UVI_RISK_VERY_HIGH:
			println("UV risk level: very high")
		case veml6070.UVI_RISK_EXTREME:
			println("UV risk level: extreme")
		}

		time.Sleep(2 * time.Second)
	}
}
