//go:build none

package main

import (
	"bytes"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
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
//
// Here are two important resources. For the timings:
// https://wp.josh.com/2014/05/13/ws2812-neopixels-are-not-so-finicky-once-you-get-to-know-them/
// For the assembly (more or less):
// https://cpldcpu.wordpress.com/2014/01/19/light_ws2812-library-v2-0/
// The timings deviate a little bit from the code here, but so far the timings
// from wp.josh.com seem to be fine for the ws2812.

// Architecture implementation. Describes the template and the timings of the
// blocks of instructions so that most code can remain architecture-independent.
type architectureImpl struct {
	buildTag         string
	minBaseCyclesT0H int
	maxBaseCyclesT0H int
	minBaseCyclesT1H int
	maxBaseCyclesT1H int
	minBaseCyclesTLD int
	valueTemplate    string // template for how to pass the 'c' byte to assembly
	template         string // assembly template
}

var architectures = map[string]architectureImpl{
	"cortexm": {
		// Assume that a branch is 1 to 3 cycles, no matter whether it's taken
		// or not. This is a rather conservative estimate, for Cortex-M+ for
		// example the instruction cycles are precisely known.
		buildTag:         "cortexm",
		minBaseCyclesT0H: 1 + 1 + 2, // shift + branch (not taken) + store
		maxBaseCyclesT0H: 1 + 3 + 2, // shift + branch (not taken) + store
		minBaseCyclesT1H: 1 + 1 + 2, // shift + branch (taken) + store
		maxBaseCyclesT1H: 1 + 3 + 2, // shift + branch (taken) + store
		minBaseCyclesTLD: 1 + 2 + 2, // subtraction + branch x2 + store (in next cycle)
		valueTemplate:    "(uint32_t)c << 24",
		template: `
1: @ send_bit
  str   %[maskSet], %[portSet]     @ [2]   T0H and T0L start here
  @DELAY1
  lsls  %[value], #1               @ [1]
  bcs.n 2f                         @ [1/3] skip_store
  str   %[maskClear], %[portClear] @ [2]   T0H -> T0L transition
2: @ skip_store
  @DELAY2
  str   %[maskClear], %[portClear] @ [2]   T1H -> T1L transition
  @DELAY3
  subs  %[i], #1                   @ [1]
  beq.n 3f                         @ [1/3] end
  b     1b                         @ [1/3] send_bit
3: @ end
`,
	},
	"tinygoriscv": {
		// Largely based on the SiFive FE310 CPU:
		// - stores are 1 cycle
		// - branches are 1 or 3 cycles, depending on branch prediction
		// - ALU operations are 1 cycle (as on most CPUs)
		// Hopefully this generalizes to other chips.
		buildTag:         "tinygo.riscv32",
		minBaseCyclesT0H: 1 + 1 + 1, // shift + branch (not taken) + store
		maxBaseCyclesT0H: 1 + 3 + 1, // shift + branch (not taken) + store
		minBaseCyclesT1H: 1 + 1 + 1, // shift + branch (taken) + store
		maxBaseCyclesT1H: 1 + 3 + 1, // shift + branch (taken) + store
		minBaseCyclesTLD: 1 + 1 + 1, // subtraction + branch + store (in next cycle)
		valueTemplate:    "(uint32_t)c << 23",
		template: `
1: // send_bit
  sw    %[maskSet], %[portSet]     // [1]   T0H and T0L start here
  @DELAY1
  slli  %[value], %[value], 1      // [1]   shift value left by 1
  bltz  %[value], 2f               // [1/3] skip_store
  sw    %[maskClear], %[portClear] // [1]   T0H -> T0L transition
2: // skip_store
  @DELAY2
  sw    %[maskClear], %[portClear] // [1]   T1H -> T1L transition
  @DELAY3
  addi  %[i], %[i], -1             // [1]
  bnez  %[i], 1b                   // [1/3] send_bit
`,
	},
}

func writeCAssembly(f *os.File, arch string, megahertz int) error {
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
	// The equivalent table for WS2811 LEDs would be the following:
	//   Symbol   Parameter                    Min   Typical    Max   Units
	//   T0H      0 code, high voltage time    350       500    650   ns
	//   T1H      1 code, high voltage time   1050      1200   5500   ns
	//   TLD      data, low voltage time      1150      1300   5000   ns
	//   TLL      latch, low voltage time     6000                    ns
	// Combining the two (min and max) leads to the following table:
	//   Symbol   Parameter                    Min   Typical    Max   Units
	//   T0H      0 code, high voltage time    350         -    500   ns
	//   T1H      1 code, high voltage time   1050         -   5500   ns
	//   TLD      data, low voltage time      1150         -   5000   ns
	//   TLL      latch, low voltage time     6000                    ns
	// These comined timings are used so that the ws2812 package is compatible
	// with both WS2812 and with WS2811 chips.
	// T0H is the time the pin should be high to send a "0" bit.
	// T1H is the time the pin should be high to send a "1" bit.
	// TLD is the time the pin should be low between bits.
	// TLL is the time the pin should be low to apply (latch) the new colors.
	minCyclesT0H := int(math.Ceil(0.350 / cycleTimeNS))
	maxCyclesT0H := int(math.Floor(0.500 / cycleTimeNS))
	minCyclesT1H := int(math.Ceil(1.050 / cycleTimeNS))
	maxCyclesT1H := int(math.Floor(5.500 / cycleTimeNS))
	minCyclesTLD := int(math.Ceil(1.150 / cycleTimeNS))

	// The assembly template looks something like this:
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
	archImpl, ok := architectures[arch]
	if !ok {
		return fmt.Errorf("unknown architecture: %s", arch)
	}

	// Determine number of nops for delay1. This is primarily based on the T0H
	// delay, which is relatively short (<500ns).
	delay1 := minCyclesT0H - archImpl.minBaseCyclesT0H
	if delay1 < 0 {
		// The minCyclesT0H constraint could not be satisfied. Don't insert
		// nops, in the hope that it isn't too long.
		delay1 = 0
	}
	if delay1+archImpl.maxBaseCyclesT0H > maxCyclesT0H {
		return fmt.Errorf("MCU appears to be too slow to satisfy minimum requirements for the T0H signal")
	}
	actualMinCyclesT0H := archImpl.minBaseCyclesT0H + delay1
	actualMaxCyclesT0H := archImpl.maxBaseCyclesT0H + delay1
	actualMinNanosecondsT0H := float64(actualMinCyclesT0H) / float64(megahertz) * 1000
	actualMaxNanosecondsT0H := float64(actualMaxCyclesT0H) / float64(megahertz) * 1000

	// Determine number of nops for delay2. This is delay1 plus some extra time
	// so that the pulse is long enough for T1H.
	minBaseCyclesT1H := delay1 + archImpl.minBaseCyclesT1H // delay1 + asssembly cycles
	maxBaseCyclesT1H := delay1 + archImpl.maxBaseCyclesT1H // delay1 + asssembly cycles
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
	delay3 := minCyclesTLD - archImpl.minBaseCyclesTLD
	if delay3 < 0 {
		delay3 = 0
	}
	actualMinCyclesTLD := archImpl.minBaseCyclesTLD + delay3
	actualMinNanosecondsTLD := float64(actualMinCyclesTLD) / float64(megahertz) * 1000

	// Create the Go function in a buffer. Using a buffer here to be able to
	// ignore I/O errors.
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "\n")
	fmt.Fprintf(buf, "__attribute__((always_inline))\nvoid ws2812_writeByte%d(char c, uint32_t *portSet, uint32_t *portClear, uint32_t maskSet, uint32_t maskClear) {\n", megahertz)
	fmt.Fprintf(buf, "	// Timings:\n")
	fmt.Fprintf(buf, "	// T0H: %2d - %2d cycles or %.1fns - %.1fns\n", actualMinCyclesT0H, actualMaxCyclesT0H, actualMinNanosecondsT0H, actualMaxNanosecondsT0H)
	fmt.Fprintf(buf, "	// T1H: %2d - %2d cycles or %.1fns - %.1fns\n", actualMinCyclesT1H, actualMaxCyclesT1H, actualMinNanosecondsT1H, actualMaxNanosecondsT1H)
	fmt.Fprintf(buf, "	// TLD: %2d -    cycles or %.1fns -\n", actualMinCyclesTLD, actualMinNanosecondsTLD)
	fmt.Fprintf(buf, "	uint32_t value = %s;\n", archImpl.valueTemplate)
	asm := archImpl.template
	asm = strings.TrimSpace(asm)
	asm = strings.ReplaceAll(asm, "  @DELAY1\n", strings.Repeat("  nop\n", delay1))
	asm = strings.ReplaceAll(asm, "  @DELAY2\n", strings.Repeat("  nop\n", delay2))
	asm = strings.ReplaceAll(asm, "  @DELAY3\n", strings.Repeat("  nop\n", delay3))
	asm = strings.ReplaceAll(asm, "\n", "\n\t")
	fmt.Fprintf(buf, "	char i = 8;\n")
	fmt.Fprintf(buf, "	__asm__ __volatile__(\n")
	for _, line := range strings.Split(asm, "\n") {
		fmt.Fprintf(buf, "\t\t%#v\n", line+"\n")
	}
	// Note: [value] and [i] must be input+output operands because they modify
	// the value.
	fmt.Fprintf(buf, `	: [value]"+r"(value),
	  [i]"+r"(i)
	: [maskSet]"r"(maskSet),
	  [portSet]"m"(*portSet),
	  [maskClear]"r"(maskClear),
	  [portClear]"m"(*portClear));
}
`)

	// Now write the buffer contents (with the assembly function) to a file.
	_, err := f.Write(buf.Bytes())
	return err
}

func writeGoWrapper(f *os.File, arch string, megahertz int) error {
	// Create the Go function in a buffer. Using a buffer here to be able to
	// ignore I/O errors.
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "\n")
	fmt.Fprintf(buf, "func (d Device) writeByte%d(c byte) {\n", megahertz)
	fmt.Fprintf(buf, "	portSet, maskSet := d.Pin.PortMaskSet()\n")
	fmt.Fprintf(buf, "	portClear, maskClear := d.Pin.PortMaskClear()\n")
	fmt.Fprintf(buf, "\n")
	fmt.Fprintf(buf, "	mask := interrupt.Disable()\n")
	fmt.Fprintf(buf, "	C.ws2812_writeByte%d(C.char(c), (*C.uint32_t)(unsafe.Pointer(portSet)), (*C.uint32_t)(unsafe.Pointer(portClear)), C.uint32_t(maskSet), C.uint32_t(maskClear))\n", megahertz)
	buf.WriteString(`
	interrupt.Restore(mask)
}
`)

	// Now write the buffer contents (with the assembly function) to a file.
	_, err := f.Write(buf.Bytes())
	return err
}

func main() {
	arch := flag.String("arch", "cortexm", "architecture to output to")
	flag.Parse()

	// Remaining parameters are all clock frequencies.
	var clockFrequencies []int
	for _, s := range flag.Args() {
		freq, err := strconv.Atoi(s)
		if err != nil {
			fmt.Fprintln(os.Stderr, "cannot parse frequency:", s)
			os.Exit(1)
		}
		clockFrequencies = append(clockFrequencies, freq)
	}

	f, err := os.Create("ws2812-asm_" + *arch + ".go")
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not generate WS2812 assembly code:", err)
		os.Exit(1)
	}
	defer f.Close()
	fmt.Fprintln(f, "//go:build", architectures[*arch].buildTag)
	f.WriteString(`
package ws2812

// Warning: autogenerated file. Instead of modifying this file, change
// gen-ws2812.go and run "go generate".

import "runtime/interrupt"
import "unsafe"

/*
#include <stdint.h>
`)
	for _, megahertz := range clockFrequencies {
		err := writeCAssembly(f, *arch, megahertz)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not generate WS2812 assembly code for %s and %dMHz: %s\n", *arch, megahertz, err)
			os.Exit(1)
		}
	}
	f.WriteString(`*/
import "C"
`)
	for _, megahertz := range clockFrequencies {
		err := writeGoWrapper(f, *arch, megahertz)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not generate Go wrapper: %w\n", err)
			os.Exit(1)
		}
	}
}
