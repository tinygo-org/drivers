package main

/*
	Example works with TM1638 with eight 7-segment indicators at odd addresses
	and eight LEDs at even addresses. Also eight buttons connected to first scan line.
	Configuration implemented by board MDU1093.
*/

import (
	"machine"
	"time"

	"tinygo.org/x/drivers/tm1638"
)

/*

Odd items of toShow array contain bytes for 7-segment LEDs.
Mapping for bits and segments shown bellow.

--- 0 ---
|       |
5       1
|       |
--- 6 ---
|       |
4       2
|       |
--- 3 ---.7

*/

func main() {
	// codes of numbers from 0 to 7 at odd indexes described
	toShow := []uint8{0x3f, 1, 0x06, 0, 0x5b, 1, 0x4f, 0, 0x66, 1, 0x6d, 0, 0x7d, 1, 0x07}
	// buffer for keyboard scan
	keyBuffer := [4]uint8{0, 0, 0, 0}
	// display memory buffer
	displayBuffer := make([]byte, 16)

	tm := tm1638.New(machine.D7, machine.D9, machine.D8) // strobe, clock, data
	config := tm1638.Config{Brightness: tm1638.MaxBrightness}
	tm.Configure(config)

	fill(displayBuffer, 0xFF)
	tm.WriteAt(displayBuffer, 0)
	time.Sleep(time.Second * 3)
	fill(displayBuffer, 0)

	// visualization of bit to segment mapping
	for i := uint8(0); i < 8; i++ {
		displayBuffer[uint8(i)<<1] = 1 << uint8(i)
	}
	tm.WriteAt(displayBuffer, 0)
	time.Sleep(time.Second * 3)

	// show eight numbers and light on odd LEDs
	tm.WriteAt(toShow, 0)
	time.Sleep(time.Millisecond * 1000)

	// 7 levels of brightness
	for i := uint8(0); i < 8; i++ {
		tm.SetBrightness(i)
		time.Sleep(time.Millisecond * 1000)
	}

	fill(displayBuffer, 0)
	displayBuffer[0] = 0x7F
	for i := 0; i < len(displayBuffer); i++ {
		if i > 0 {
			// move light to next position
			displayBuffer[i], displayBuffer[i-1] = displayBuffer[i-1], 0
		}
		tm.WriteAt(displayBuffer, 0)
		time.Sleep(time.Millisecond * 250)
	}

	//  7-segment indicator index
	var indicatorIndex uint8 = 0
	// index of segment to switch on
	segmentIndex := 0
	for {
		// prepare buffer
		fill(displayBuffer, 0)

		// scan pressed keys
		tm.ScanKeyboard(&keyBuffer)

		// translate scan result into bits of one uint8
		var firstScanLine uint8 = 0
		for i := 0; i < len(keyBuffer); i++ {
			firstScanLine |= (keyBuffer[i] << i)
		}

		// switch LEDs on above the pressed buttons
		for i := 0; i < 8; i++ {
			if (firstScanLine & (1 << i)) > 0 {
				// LED switch on
				displayBuffer[1+uint8(i)<<1] = 0xFF
			}
		}

		// next segment switch on
		displayBuffer[indicatorIndex<<1] = 1 << segmentIndex

		segmentIndex++
		if segmentIndex == 8 {
			segmentIndex = 0
			indicatorIndex++
		}
		if indicatorIndex == 8 {
			indicatorIndex = 0
			segmentIndex = 0
		}

		tm.WriteAt(displayBuffer, 0)
		time.Sleep(time.Millisecond * 50)
	}
}

func fill(values []byte, value byte) {
	for i, _ := range values {
		values[i] = value
	}
}
