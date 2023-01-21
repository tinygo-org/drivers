// Package gps provides a driver for GPS receivers over UART and I2C
package gps // import "tinygo.org/x/drivers/gps"

import (
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"tinygo.org/x/drivers"
)

var (
	errInvalidNMEASentenceLength = errors.New("invalid NMEA sentence length")
	errInvalidNMEAChecksum       = errors.New("invalid NMEA sentence checksum")
	errEmptyNMEASentence         = errors.New("cannot parse empty NMEA sentence")
	errUnknownNMEASentence       = errors.New("unsupported NMEA sentence type")
	errInvalidGGASentence        = errors.New("invalid GGA NMEA sentence")
	errInvalidRMCSentence        = errors.New("invalid RMC NMEA sentence")
	errInvalidGLLSentence        = errors.New("invalid GLL NMEA sentence")
)

type GPSError struct {
	Err      error
	Info     string
	Sentence string
}

func newGPSError(err error, sentence string, info string) GPSError {
	return GPSError{
		Info:     info,
		Err:      err,
		Sentence: sentence,
	}
}

func (ge GPSError) Error() string {
	return ge.Err.Error() + " " + ge.Info + " " + ge.Sentence
}

func (ge GPSError) Unwrap() error {
	return ge.Err
}

const (
	minimumNMEALength = 7
	startingDelimiter = '$'
	checksumDelimiter = '*'
)

// Device wraps a connection to a GPS device.
type Device struct {
	buffer   []byte
	bufIdx   int
	sentence strings.Builder
	uart     drivers.UART
	bus      drivers.I2C
	address  uint16
}

// NewUART creates a new UART GPS connection. The UART must already be configured.
func NewUART(uart drivers.UART) Device {
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

	for b != startingDelimiter {
		b = gps.readNextByte()
	}

	for b != checksumDelimiter {
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
// For example, a valid NMEA sentence such as this:
// $GPGLL,3751.65,S,14507.36,E*77
// It has to start with a '$' character.
// It has to have a 5 character long sentence identifier.
// It has to end with a '*' character following by a checksum.
func validSentence(sentence string) error {
	if len(sentence) < minimumNMEALength || sentence[0] != startingDelimiter || sentence[len(sentence)-3] != checksumDelimiter {
		return errInvalidNMEASentenceLength
	}
	var cs byte = 0
	for i := 1; i < len(sentence)-3; i++ {
		cs ^= sentence[i]
	}
	checksum := strings.ToUpper(hex.EncodeToString([]byte{cs}))
	if checksum != sentence[len(sentence)-2:len(sentence)] {
		return newGPSError(errInvalidNMEAChecksum, sentence,
			"expected "+sentence[len(sentence)-2:len(sentence)]+
				" got "+checksum)
	}

	return nil
}
