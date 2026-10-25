// Package ws2812 implements a driver for WS2812 and SK6812 RGB LED strips.
//
// On most platforms NewWS2812 uses bit-banging.
// On RP2040/RP2350 it uses PIO for hardware-timed control.
package ws2812 // import "tinygo.org/x/drivers/ws2812"

//go:generate go run gen-ws2812.go -arch=cortexm 16 48 64 120 125 150 160 168 200
//go:generate go run gen-ws2812.go -arch=tinygoriscv 320
//go:generate go run gen-ws2812.go -arch=esp32c3 160
