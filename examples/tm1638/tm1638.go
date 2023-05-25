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

func main() {
	// codes of nombers from 0 to 7 at odd indexes
	toShow := []uint8{0x3f, 1, 0x06, 0, 0x5b, 1, 0x4f, 0, 0x66, 1, 0x6d, 0, 0x7d, 1, 0x07}
	// buffer for keyboard scan
	var keyBuffer = [4]uint8{0, 0, 0, 0}

	tm := tm1638.New(machine.D7, machine.D9, machine.D8) // strobe, clock, data
	tm.Configure()
	// show eight numbers and light on odd LEDs
	tm.WriteArray(0, toShow)
	time.Sleep(time.Millisecond * 1000)

	// 7 levels of brightness
	for i := uint8(0); i < 8; i++ {
		tm.Brightness(i)
		time.Sleep(time.Millisecond * 1000)
	}
	tm.Clear()

	// light on and off each indicator and LED
	for i := uint8(0); i < 16; i++ {
		if i > 0 {
			//
			tm.Write(i-1, 0x00)
		}
		tm.Write(i, 0x7F)
		time.Sleep(time.Millisecond * 250)
	}

	//  7-segment indicator index
	var indicatorIndex uint8 = 0
	// index of segment to switch on
	segmentIndex := 0

	for {
		// scan pressed keys
		tm.ScanKeyboard(&keyBuffer)

		// translate scan result into bits of one uint8
		var firstScanLine uint8 = 0
		for i := 0; i < len(keyBuffer); i++ {
			firstScanLine |= (keyBuffer[i] << i)
		}

		// i is index of button
		for i := 0; i < 8; i++ {
			if (firstScanLine & (1 << i)) > 0 {
				// switch on LED
				tm.Write(1+uint8(i)<<1, 0xff)
			} else {
				// switch off LEd
				tm.Write(1+uint8(i)<<1, 0x00)
			}
		}

		// switch off all segments
		tm.Write(indicatorIndex<<1, 0x00)
		if segmentIndex == 8 {
			segmentIndex = 0
			indicatorIndex++
		}
		if indicatorIndex == 9 {
			indicatorIndex = 0
		}
		// next segment switch on
		tm.Write(indicatorIndex<<1, 1<<segmentIndex)
		segmentIndex++

		time.Sleep(time.Millisecond * 50)
	}
}
