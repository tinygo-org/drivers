package tm1638

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
	var keyBuffer = [4]uint8{0, 0, 0, 0}

	tm := tm1638.New(machine.D7, machine.D9, machine.D8) // strobe, clock, data
	config := tm1638.Config{Brightness: tm1638.MaxBrightness}
	tm.Configure(config)

	// visualization of bit to segment mapping
	for i := uint8(0); i < 8; i++ {
		tm.Write(1<<uint8(i), uint8(i)<<1)
	}
	time.Sleep(time.Second * 3)

	// show eight numbers and light on odd LEDs
	tm.WriteAt(toShow, 0)
	time.Sleep(time.Millisecond * 1000)

	// 7 levels of brightness
	for i := uint8(0); i < 8; i++ {
		tm.SetBrightness(i)
		time.Sleep(time.Millisecond * 1000)
	}
	tm.Clear()

	// light on and off each indicator and LED
	for i := uint8(0); i < 16; i++ {
		if i > 0 {
			//
			tm.Write(0x00, i-1)
		}
		tm.Write(0x7F, i)
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
				// LED switch on
				tm.Write(0xff, 1+uint8(i)<<1)
			} else {
				// LED switch off
				tm.Write(0x00, 1+uint8(i)<<1)
			}
		}

		// switch off all segments
		tm.Write(0x00, indicatorIndex<<1)
		if segmentIndex == 8 {
			segmentIndex = 0
			indicatorIndex++
		}
		if indicatorIndex == 9 {
			indicatorIndex = 0
		}
		// next segment switch on
		tm.Write(1<<segmentIndex, indicatorIndex<<1)
		segmentIndex++

		time.Sleep(time.Millisecond * 50)
	}
}
