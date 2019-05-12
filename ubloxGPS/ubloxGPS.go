// Package ubloxGPS provides a driver for UBlox GPS receivers over I2C
//
// Datasheet:
// https://www.u-blox.com/sites/default/files/products/documents/u-blox8-M8_ReceiverDescrProtSpec_%28UBX-13003221%29_Public.pdf
// (Section 11.5)
//
package ubloxGPS

import (
	"machine"
	"time"
)

// Device wraps an I2C connection to a ublox gps device.
type GPSDevice struct {
	bus        machine.I2C
	Address    uint16
	buffer     []byte
	sentence   []byte
	ringBuffer *machine.RingBuffer
}

// New creates a new GPS connection. The I2C bus must already be
// configured.
//
// This function only creates the GPSDevice object, it does not initialize the device.
// You must call Configure() first in order to use the device itself.
func New(bus machine.I2C) GPSDevice {
	return GPSDevice{
		bus:     bus,
		Address: Address,
		// TODO: bit rubbish having 3 of these buffer type things
		buffer:     make([]byte, buffer_size),
		sentence:   make([]byte, 128), //? enough for the longest single sentence
		ringBuffer: machine.NewRingBuffer(),
	}
}

// Available returns how many bytes of GPS data are currently available.
func (gps *GPSDevice) available() (available int) {
	var lengthBytes [2]byte
	gps.bus.Tx(gps.Address, []byte{FD}, lengthBytes[0:2])
	available = int(lengthBytes[0])*256 + int(lengthBytes[1])
	return available
}

func (gps *GPSDevice) read() {
	// println("read")

	var available int
	for {
		available = gps.available()
		if available > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	var bytesToRead = min(available, int(buffer_size-gps.ringBuffer.Used()))
	gps.bus.Tx(gps.Address, []byte{FF}, gps.buffer[0:bytesToRead])
	for i := 0; i < bytesToRead; i += 1 {
		gps.ringBuffer.Put(gps.buffer[i])
	}
	// print("[[[")
	// print(string(gps.buffer[0:bytesToRead]))
	// println("]]]")
}

func (gps *GPSDevice) readNextByte() (b byte) {
	for {
		if gps.ringBuffer.Used() == 0 {
			gps.read()
		}
		var b, _ = gps.ringBuffer.Get()
		return b
	}
}

func (gps *GPSDevice) readToNextDollar() (b byte) {
	for {
		var b = gps.readNextByte()
		if b == '$' {
			return b
		}
	}
}

func (gps *GPSDevice) ReadNextSentence() (sentence string) {
	// println("ReadNextSentence")
	var i = 0
	gps.sentence[i] = gps.readToNextDollar()
	for {
		i += 1
		var b = gps.readNextByte()
		gps.sentence[i] = b
		if b == '*' {
			return string(gps.sentence[0 : i+1])
		}
	}
}
