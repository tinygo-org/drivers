package main

import (
	"time"

	"tinygo.org/x/drivers/delay"
)

func main() {
	time.Sleep(time.Second) // wait for a serial console
	start := time.Now()
	for i := 0; i < 2000; i++ {
		delay.Sleep(50 * time.Microsecond)
	}
	duration := time.Since(start)
	println("sleep of 2000*50µs (100ms) took:", duration.String())
}
