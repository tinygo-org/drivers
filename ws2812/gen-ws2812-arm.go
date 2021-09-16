//go:build none
// +build none

package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strings"
)

// This file generates assembly to precisely time the WS2812 protocol for
// various chips. Just add a new frequency below and run `go generate` to add
// the new assembly implementation - no fiddly timings to calculate and no nops
// to count!
//
// Right now this is specific to Cortex-M chips and assume the following things:
// - Arithmetic operations (shift, add, sub) take up 1 clock cycle.
// - The nop instruction also takes up 1 clock cycle.
// - Store instructions (to the GPIO pins) take up 2 clock cycles.
// - Branch instructions can take up 1 to 3 clock cycles. On the Cortex-M0, this
//   depends on whether the branch is taken or not. On the M4, the documentation
//   is less clear but it appears the instruction is still 1 to 3 cycles
//   (possibly including some branch prediction).
// It is certainly possible to extend this to other architectures, such as AVR
// and RISC-V if needed.

// Clock frequencies to support, in MHz.
var clockFrequencies = []int{16, 48, 64, 120}

func writeImplementation(f *os.File, megahertz int) error {
	cycleTimeNS := 1 / float64(megahertz)
	// These timings are taken from the table "Updated simplified timing
	// constraints for NeoPixel strings" at:
	// https://wp.josh.com/2014/05/13/ws2812-neopixels-are-not-so-finicky-once-you-get-to-know-them/
	// Here is a copy:
	//   Symbol   Parameter                    Min   Typical    Max   Units
	//   T0H      0 code, high voltage time    200       350    500   ns
	//   T1H      1 code, high voltage time    550       700   5500   ns
	//   TLD      data, low voltage time       450       600   5000   ns
	//   TLL      latch, low voltage time     6000                    ns
	// T0H is the time the pin should be high to send a "0" bit.
	// T1H is the time the pin should be high to send a "1" bit.
	// TLD is the time the pin should be low between bits.
	// TLL is the time the pin should be low to apply (latch) the new colors.
	minCyclesT0H := int(math.Ceil(0.200 / cycleTimeNS))
	maxCyclesT0H := int(math.Floor(0.500 / cycleTimeNS))
	minCyclesT1H := int(math.Ceil(0.550 / cycleTimeNS))
	maxCyclesT1H := int(math.Floor(5.500 / cycleTimeNS))
	minCyclesTLD := int(math.Ceil(0.450 / cycleTimeNS))

	// Assembly template:
	// 1: @ send_bit
	//   str   {maskSet}, {portSet}     @ [2]   T0H and T0L start here
	//   ...delay 1
	//   lsls  {value}, #1              @ [1]
	//   bcs.n 2f                       @ [1/3] skip_store
	//   str   {maskClear}, {portClear} @ [2]   T0H -> T0L transition
	// 2: @ skip_store
	//   ...delay 2
	//   str   {maskClear}, {portClear} @ [2]   T1H -> T1L transition
	//   ...delay 3
	//   subs  {i}, #1                  @ [1]
	//   bne.n 1b                       @ [1/3] send_bit
	//
	// We need to calculate the number of nop instructions in the three delays.

	// Determine number of nops for delay1. This is primarily based on the T0H
	// delay, which is relatively short (<500ns).
	minBaseCyclesT0H := 1 + 1 + 2 // shift + branch + store
	maxBaseCyclesT0H := 1 + 3 + 2 // shift + branch + store
	delay1 := minCyclesT0H - minBaseCyclesT0H
	if delay1 < 0 {
		// The minCyclesT0H constraint could not be satisfied. Don't insert
		// nops, in the hope that it isn't too long.
		delay1 = 0
	}
	if delay1+maxBaseCyclesT0H > maxCyclesT0H {
		return fmt.Errorf("MCU appears to be too slow to satisfy minimum requirements for the T0H signal")
	}
	actualMinCyclesT0H := minBaseCyclesT0H + delay1
	actualMaxCyclesT0H := maxBaseCyclesT0H + delay1
	actualMinNanosecondsT0H := float64(actualMinCyclesT0H) / float64(megahertz) * 1000
	actualMaxNanosecondsT0H := float64(actualMaxCyclesT0H) / float64(megahertz) * 1000

	// Determine number of nops for delay2. This is delay1 plus some extra time
	// so that the pulse is long enough for T1H.
	minBaseCyclesT1H := delay1 + 1 + 1 + 2 // delay1 + shift + branch + store
	maxBaseCyclesT1H := delay1 + 1 + 3 + 2 // delay1 + shift + branch + store
	delay2 := minCyclesT1H - minBaseCyclesT1H
	if delay2 < 0 {
		delay2 = 0
	}
	if delay2+maxBaseCyclesT1H > maxCyclesT1H {
		// Unlikely, we have 5500ns for this operation.
		return fmt.Errorf("MCU appears to be too slow to satisfy minimum requirements for the T1H signal")
	}
	actualMinCyclesT1H := minBaseCyclesT1H + delay2
	actualMaxCyclesT1H := maxBaseCyclesT1H + delay2
	actualMinNanosecondsT1H := float64(actualMinCyclesT1H) / float64(megahertz) * 1000
	actualMaxNanosecondsT1H := float64(actualMaxCyclesT1H) / float64(megahertz) * 1000

	// Determine number of nops for delay3. This is based on the TLD delay, the
	// time between two high pulses.
	minBaseCyclesTLD := 1 + 1 + 2 // subtraction + branch + store (in next cycle)
	delay3 := minCyclesTLD - minBaseCyclesTLD
	if delay3 < 0 {
		delay3 = 0
	}
	actualMinCyclesTLD := minBaseCyclesTLD + delay3
	actualMinNanosecondsTLD := float64(actualMinCyclesTLD) / float64(megahertz) * 1000

	// Create the Go function in a buffer. Using a buffer here to be able to
	// ignore I/O errors.
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "\n")
	fmt.Fprintf(buf, "func (d Device) writeByte%d(c byte) {\n", megahertz)
	fmt.Fprintf(buf, "	portSet, maskSet := d.Pin.PortMaskSet()\n")
	fmt.Fprintf(buf, "	portClear, maskClear := d.Pin.PortMaskClear()\n")
	fmt.Fprintf(buf, "\n")
	fmt.Fprintf(buf, "	// Timings:\n")
	fmt.Fprintf(buf, "	// T0H: %2d - %2d cycles or %.1fns - %.1fns\n", actualMinCyclesT0H, actualMaxCyclesT0H, actualMinNanosecondsT0H, actualMaxNanosecondsT0H)
	fmt.Fprintf(buf, "	// T1H: %2d - %2d cycles or %.1fns - %.1fns\n", actualMinCyclesT1H, actualMaxCyclesT1H, actualMinNanosecondsT1H, actualMaxNanosecondsT1H)
	fmt.Fprintf(buf, "	// TLD: %2d -    cycles or %.1fns -\n", actualMinCyclesTLD, actualMinNanosecondsTLD)
	fmt.Fprintf(buf, "	mask := interrupt.Disable()\n")
	fmt.Fprintf(buf, "	value := uint32(c) << 24\n")
	fmt.Fprintf(buf, "	device.AsmFull(`\n")
	fmt.Fprintf(buf, "	1: @ send_bit\n")
	fmt.Fprintf(buf, "	  str   {maskSet}, {portSet}     @ [2]   T0H and T0L start here\n")
	buf.WriteString(strings.Repeat("	  nop\n", delay1))
	fmt.Fprintf(buf, "	  lsls  {value}, #1              @ [1]\n")
	fmt.Fprintf(buf, "	  bcs.n 2f                       @ [1/3] skip_store\n")
	fmt.Fprintf(buf, "	  str   {maskClear}, {portClear} @ [2]   T0H -> T0L transition\n")
	fmt.Fprintf(buf, "	2: @ skip_store\n")
	buf.WriteString(strings.Repeat("	  nop\n", delay2))
	fmt.Fprintf(buf, "	  str   {maskClear}, {portClear} @ [2]   T1H -> T1L transition\n")
	buf.WriteString(strings.Repeat("	  nop\n", delay3))
	fmt.Fprintf(buf, "	  subs  {i}, #1                  @ [1]\n")
	fmt.Fprintf(buf, "	  bne.n 1b                       @ [1/3] send_bit\n")
	fmt.Fprintf(buf, "	`, map[string]interface{}{")
	buf.WriteString(`
		"value":     value,
		"i":         8,
		"maskSet":   maskSet,
		"portSet":   portSet,
		"maskClear": maskClear,
		"portClear": portClear,
	})
	interrupt.Restore(mask)
}
`)

	// Now write the buffer contents (with the assembly function) to a file.
	_, err := f.Write(buf.Bytes())
	return err
}

func main() {
	f, err := os.Create("ws2812-asm_cortexm.go")
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not generate WS2812 assembly code:", err)
		os.Exit(1)
	}
	defer f.Close()
	f.WriteString(`// +build cortexm

package ws2812

// Warning: autogenerated file. Instead of modifying this file, change
// gen-ws2812-arm.go and run "go generate".

import (
	"device"
	"runtime/interrupt"
)
`)
	for _, megahertz := range clockFrequencies {
		err := writeImplementation(f, megahertz)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not generate WS2812 assembly code for %dMHz: %s\n", megahertz, err)
			os.Exit(1)
		}
	}
}
