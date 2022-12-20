//go:build nano_rp2040

// This examples shows how to control RGB LED connected to
// NINA-W102 chip on Arduino Nano RP2040 Connect board
// Built-in LED code added for API comparison

package main

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/wifinina"
)

const (
	LED = machine.LED

	// Arduino Nano RP2040 Connect board RGB LED pins
	// See https://docs.arduino.cc/static/3525d638b5c76a2d19588d6b41cd02a0/ABX00053-full-pinout.pdf
	LED_R wifinina.Pin = 27
	LED_G wifinina.Pin = 25
	LED_B wifinina.Pin = 26
)

var (

	// these are the default pins for the Arduino Nano-RP2040 Connect
	spi = machine.NINA_SPI

	// this is the ESP chip that has the WIFININA firmware flashed on it
	device *wifinina.Device
)

func setup() {

	// Configure SPI for 8Mhz, Mode 0, MSB First
	spi.Configure(machine.SPIConfig{
		Frequency: 8 * 1e6,
		SDO:       machine.NINA_SDO,
		SDI:       machine.NINA_SDI,
		SCK:       machine.NINA_SCK,
	})

	device = wifinina.New(spi,
		machine.NINA_CS,
		machine.NINA_ACK,
		machine.NINA_GPIO0,
		machine.NINA_RESETN)
	device.Configure()

	time.Sleep(time.Second)

	LED.Configure(machine.PinConfig{Mode: machine.PinOutput})
	LED_R.Configure(wifinina.PinConfig{Mode: wifinina.PinOutput})
	LED_G.Configure(wifinina.PinConfig{Mode: wifinina.PinOutput})
	LED_B.Configure(wifinina.PinConfig{Mode: wifinina.PinOutput})
}

func main() {

	setup()

	LED.Low()    // OFF
	LED_R.High() // OFF
	LED_G.High() // OFF
	LED_B.High() // OFF

	go func() {
		for {
			LED.Low()
			time.Sleep(time.Second)
			LED.High()
			time.Sleep(time.Second)
		}
	}()

	for {
		LED_R.Low() // ON
		time.Sleep(time.Second)
		LED_R.High() // OFF
		LED_G.Low()  // ON
		time.Sleep(time.Second)
		LED_G.High() // OFF
		LED_B.Low()  // ON
		time.Sleep(time.Second)
		LED_B.High() // OFF
	}

}
