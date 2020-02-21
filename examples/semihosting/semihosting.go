package main

// A small example that demonstrates how SemiHosting can be used.
// You could use it with a board that supports GDB, such as the BBC micro:bit:
// 1. Compile and debug it:
//      tinygo gdb -target=microbit -ocd-output tinygo.org/x/drivers/examples/semihosting
// 2. Enable semihosting in the GDB shell:
//      monitor arm semihosting enable
// 3. Start the program:
//      continue

import (
	"time"

	"tinygo.org/x/drivers/semihosting"
)

func main() {
	for {
		semihosting.Stdout.Write([]byte("hello world!\n"))
		time.Sleep(time.Second)
	}
}
