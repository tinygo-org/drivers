// Package ubloxGPS provides a driver for UBlox GPS receivers over I2C
//
// Datasheet:
// https://www.u-blox.com/sites/default/files/products/documents/u-blox8-M8_ReceiverDescrProtSpec_%28UBX-13003221%29_Public.pdf
// (Section 11.5)
//
package ubloxgps

import (
	"machine"
	"strings"
	"time"
)

// Device wraps an I2C connection to a ublox gps device.
type GPSDevice struct {
	bus      machine.I2C
	Address  uint16
	buffer   []byte
	bufIdx   int
	sentence strings.Builder
}

// New creates a new GPS connection. The I2C bus must already be configured.
func New(bus machine.I2C) GPSDevice {
	return GPSDevice{
		bus:      bus,
		Address:  Address,
		buffer:   make([]byte, buffer_size),
		bufIdx:   buffer_size,
		sentence: strings.Builder{},
	}
}

// ReadNextSentence returns the next NMEA sentence from the GPS device.
func (gps *GPSDevice) ReadNextSentence() (sentence string) {
	// println("ReadNextSentence")
	gps.sentence.Reset()
	var b byte = ' '

	for b != '$' {
		b = gps.readNextByte()
	}

	for b != '*' {
		gps.sentence.WriteByte(b)
		b = gps.readNextByte()
	}
	gps.sentence.WriteByte(b)
	gps.sentence.WriteByte(gps.readNextByte())
	gps.sentence.WriteByte(gps.readNextByte())

	sentence = gps.sentence.String()
	// print(">>")
	// print(sentence)
	// println("<<")
	return sentence
}

func (gps *GPSDevice) readNextByte() (b byte) {
	gps.bufIdx += 1
	if gps.bufIdx >= buffer_size {
		gps.fillBuffer()
	}
	return gps.buffer[gps.bufIdx]
}

func (gps *GPSDevice) fillBuffer() {
	// println("read")

	for gps.available() < buffer_size {
		time.Sleep(100 * time.Millisecond)
	}

	gps.bus.Tx(gps.Address, []byte{FF}, gps.buffer[0:buffer_size])
	gps.bufIdx = 0

	// print("[[[")
	// print(string(gps.buffer[0:bytesToRead]))
	// println("]]]")
}

// Available returns how many bytes of GPS data are currently available.
func (gps *GPSDevice) available() (available int) {
	var lengthBytes [2]byte
	gps.bus.Tx(gps.Address, []byte{FD}, lengthBytes[0:2])
	available = int(lengthBytes[0])*256 + int(lengthBytes[1])
	return available
}
