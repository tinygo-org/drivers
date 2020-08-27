// Package gps provides a driver for GPS receivers over UART and I2C
package gps // import "tinygo.org/x/drivers/gps"

import (
	"encoding/hex"
	"errors"
	"machine"
	"strings"
	"time"

	"tinygo.org/x/drivers"
)

var (
	errInvalidNMEASentenceLength = errors.New("invalid NMEA sentence length")
	errInvalidNMEAChecksum       = errors.New("invalid NMEA sentence checksum")
)

// Device wraps a connection to a GPS device.
type Device struct {
	buffer   []byte
	bufIdx   int
	sentence strings.Builder
	uart     *machine.UART
	bus      drivers.I2C
	address  uint16
}

// NewUART creates a new UART GPS connection. The UART must already be configured.
func NewUART(uart *machine.UART) Device {
	return Device{
		uart:     uart,
		buffer:   make([]byte, bufferSize),
		bufIdx:   bufferSize,
		sentence: strings.Builder{},
	}
}

// NewI2C creates a new I2C GPS connection.
func NewI2C(bus drivers.I2C) Device {
	return Device{
		bus:      bus,
		address:  I2C_ADDRESS,
		buffer:   make([]byte, bufferSize),
		bufIdx:   bufferSize,
		sentence: strings.Builder{},
	}
}

// NextSentence returns the next valid NMEA sentence from the GPS device.
func (gps *Device) NextSentence() (sentence string, err error) {
	sentence = gps.readNextSentence()
	if err = validSentence(sentence); err != nil {
		return "", err
	}
	return sentence, nil
}

// readNextSentence returns the next sentence from the GPS device.
func (gps *Device) readNextSentence() (sentence string) {
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

func (gps *Device) readNextByte() (b byte) {
	gps.bufIdx += 1
	if gps.bufIdx >= bufferSize {
		gps.fillBuffer()
	}
	return gps.buffer[gps.bufIdx]
}

func (gps *Device) fillBuffer() {
	if gps.uart != nil {
		gps.uartFillBuffer()
	} else {
		gps.i2cFillBuffer()
	}
}

func (gps *Device) uartFillBuffer() {
	for gps.uart.Buffered() < bufferSize {
		time.Sleep(100 * time.Millisecond)
	}
	gps.uart.Read(gps.buffer[0:bufferSize])
	gps.bufIdx = 0
}

func (gps *Device) i2cFillBuffer() {
	for gps.available() < bufferSize {
		time.Sleep(100 * time.Millisecond)
	}
	gps.bus.Tx(gps.address, []byte{DATA_STREAM_REG}, gps.buffer[0:bufferSize])
	gps.bufIdx = 0
}

// Available returns how many bytes of GPS data are currently available.
func (gps *Device) available() (available int) {
	var lengthBytes [2]byte
	gps.bus.Tx(gps.address, []byte{BYTES_AVAIL_REG}, lengthBytes[0:2])
	available = int(lengthBytes[0])*256 + int(lengthBytes[1])
	return available
}

// WriteBytes sends data/commands to the GPS device
func (gps *Device) WriteBytes(bytes []byte) {
	if gps.uart != nil {
		gps.uart.Write(bytes)
	} else {
		gps.bus.Tx(gps.address, []byte{}, bytes)
	}
}

// validSentence checks if a sentence has been received uncorrupted
func validSentence(sentence string) error {
	if len(sentence) < 4 || sentence[0] != '$' || sentence[len(sentence)-3] != '*' {
		return errInvalidNMEASentenceLength
	}
	var cs byte = 0
	for i := 1; i < len(sentence)-3; i++ {
		cs ^= sentence[i]
	}
	checksum := hex.EncodeToString([]byte{cs})
	if (checksum[0] != sentence[len(sentence)-2]) || (checksum[1] != sentence[len(sentence)-1]) {
		return errInvalidNMEAChecksum
	}

	return nil
}
