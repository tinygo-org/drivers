// Package gps provides a driver for GPS receivers over UART and I2C
package gps

import (
	"machine"
	"strings"
	"time"
)

// Device wraps a connection to a GPS device.
type GPSDevice struct {
	buffer   []byte
	bufIdx   int
	sentence strings.Builder
	uart     machine.UART
	bus      machine.I2C
	address  uint16
}

// New creates a new UART GPS connection. The UART must already be configured.
func New(uart machine.UART) GPSDevice {
	return GPSDevice{
		uart:     uart,
		buffer:   make([]byte, bufferSize),
		bufIdx:   bufferSize,
		sentence: strings.Builder{},
	}
}

// NewI2C creates a new I2C GPS connection.
func NewI2C(bus machine.I2C) GPSDevice {
	return GPSDevice{
		bus:      bus,
		address:  I2C_ADDRESS,
		buffer:   make([]byte, bufferSize),
		bufIdx:   bufferSize,
		sentence: strings.Builder{},
	}
}

// ReadNextSentence returns the next NMEA sentence from the GPS device.
func (gps *GPSDevice) ReadNextSentence() (sentence string) {
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
	return sentence
}

func (gps *GPSDevice) readNextByte() (b byte) {
	gps.bufIdx += 1
	if gps.bufIdx >= bufferSize {
		gps.fillBuffer()
	}
	return gps.buffer[gps.bufIdx]
}

func (gps *GPSDevice) fillBuffer() {
	if (machine.I2C{}) == gps.bus {
		gps.uartFillBuffer()
	} else {
		gps.i2cFillBuffer()
	}
	// print("[[[")
	// print(string(gps.buffer[0:bufferSize]))
	// println("]]]")
}

func (gps *GPSDevice) uartFillBuffer() {
	for gps.uart.Buffered() < bufferSize {
		time.Sleep(100 * time.Millisecond)
	}
	gps.uart.Read(gps.buffer[0:bufferSize])
	gps.bufIdx = 0
}

func (gps *GPSDevice) i2cFillBuffer() {
	for gps.available() < bufferSize {
		time.Sleep(100 * time.Millisecond)
	}
	gps.bus.Tx(gps.address, []byte{DATA_STREAM_REG}, gps.buffer[0:bufferSize])
	gps.bufIdx = 0
}

// Available returns how many bytes of GPS data are currently available.
func (gps *GPSDevice) available() (available int) {
	var lengthBytes [2]byte
	gps.bus.Tx(gps.address, []byte{BYTES_AVAIL_REG}, lengthBytes[0:2])
	available = int(lengthBytes[0])*256 + int(lengthBytes[1])
	return available
}
