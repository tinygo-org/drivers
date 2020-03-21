package main

import (
	"machine"
	"time"
	"tinygo.org/x/drivers/max7219"
)

var X = [8]byte{
	0b10000001,
	0b01000010,
	0b00100100,
	0b00011000,
	0b00011000,
	0b00100100,
	0b01000010,
	0b10000001,
}

var PLAY = [8]byte{
	0b00000000,
	0b01100000,
	0b01111000,
	0b01111110,
	0b01111111,
	0b01111110,
	0b01111000,
	0b01100000,
}

var PAUSE = [8]byte{
	0b00000000,
	0b01100110,
	0b01100110,
	0b01100110,
	0b01100110,
	0b01100110,
	0b01100110,
	0b00000000,
}

var STOP = [8]byte{
	0b00000000,
	0b00000000,
	0b00111100,
	0b00111100,
	0b00111100,
	0b00111100,
	0b00000000,
	0b00000000,
}

var PICTURES = [4][8]byte{
	X,
	PLAY,
	PAUSE,
	STOP,
}

// Draw to the matix: an X, the play symbol, pause symbol, and stop symbol (square)
func main() {
	ma := &max7219.Device{machine.Pin(2), machine.Pin(3), machine.Pin(4), 1}
	ma.Configure()
	for _, pic := range PICTURES {
		ma.WriteMatrix(&pic)
		time.Sleep(time.Second)
	}
}
