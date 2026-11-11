//go:build esp32s3

package main

import "machine"

// Pins for an ESP32-S3 devkit with the WeAct module on GPIO9 to GPIO14.
const (
	blPin  = machine.GPIO9
	rstPin = machine.GPIO10
	dcPin  = machine.GPIO11
	csPin  = machine.GPIO12
	sckPin = machine.GPIO13
	sdoPin = machine.GPIO14
)
