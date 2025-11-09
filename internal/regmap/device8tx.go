// Package regmap provides transaction-based interfaces for reading and writing
// to device registers over I2C and SPI buses with pre-allocated buffers.
package regmap

import (
	"errors"

	"tinygo.org/x/drivers"
)

var (
	// errNotInTx indicates an operation was attempted outside of an active transaction.
	errNotInTx = errors.New("device not in Tx")

	// errInTx indicates a transaction was started while another is still active.
	errInTx = errors.New("device already in Tx")

	// errShortWriteBuffer indicates the write buffer is too small for the requested operation.
	errShortWriteBuffer = errors.New("device write buffer too short")

	// errShortReadBuffer indicates the read buffer is too small for the requested operation.
	errShortReadBuffer = errors.New("device read buffer too short")
)

// Device8Txer wraps a Device8 to provide buffered transaction support for
// I2C and SPI operations. It maintains pre-allocated buffers to avoid heap
// allocations during register access operations.
//
// Users must call SetBuffers to configure the write and read buffers before
// initiating transactions.
type Device8Txer struct {
	Device8
	writeBuf []byte // Pre-allocated buffer for write operations
	readBuf  []byte // Pre-allocated buffer for read operations
	inTx     bool   // Tracks whether a transaction is currently active
}

// SetBuffers configures the write and read buffers for this device.
// These buffers are reused across transactions to avoid heap allocations.
//
// The writebuf should be large enough to hold the register address plus
// all data bytes to be written in a single transaction.
func (d *Device8Txer) SetBuffers(writebuf, readbuf []byte) {
	d.readBuf = readbuf
	d.writeBuf = writebuf
}

// Tx8 represents an active transaction for an 8-bit register device.
// It tracks the write buffer and current offset as data is added to the transaction.
//
// Use AddWriteByte or AddWriteData to add data to the transaction, then call
// DoTxI2C or DoTxSPI to execute the transaction over the bus.
type Tx8 struct {
	dw  *Device8Txer // Reference to the parent device
	off int          // Current offset in the write buffer
}

// Tx initiates a new transaction for writing to the specified register address.
//
// Parameters:
//   - writeAddr: The 8-bit register address to write to
//
// Returns a Tx8 handle that can be used to add data and execute the transaction.
//
// Returns an error if:
//   - A transaction is already active (errInTx)
//   - The write buffer is too short (errShortWriteBuffer)
func (dw *Device8Txer) Tx(writeAddr uint8) (Tx8, error) {
	if dw.inTx {
		return Tx8{}, errInTx
	} else if len(dw.writeBuf) < 1 {
		return Tx8{}, errShortWriteBuffer
	}
	dw.writeBuf[0] = writeAddr
	return Tx8{dw: dw, off: 1}, nil
}

// AddWriteData appends multiple bytes to the current transaction's write buffer.
//
// Parameters:
//   - buf: Variable number of bytes to add to the transaction
//
// Returns an error if:
//   - No transaction is active (errNotInTx)
//   - The write buffer doesn't have enough space (errShortWriteBuffer)
func (tx *Tx8) AddWriteData(buf ...byte) error {
	if !tx.dw.inTx {
		return errNotInTx
	}
	avail := tx.dw.writeBuf[tx.off:]
	if len(avail) < len(buf) {
		return errShortWriteBuffer
	}
	n := copy(avail, buf)
	tx.off += n
	return nil
}

// AddWriteByte appends a single byte to the current transaction's write buffer.
//
// Parameters:
//   - b: The byte to add to the transaction
//
// Returns an error if:
//   - No transaction is active (errNotInTx)
//   - The write buffer doesn't have enough space (errShortWriteBuffer)
func (tx *Tx8) AddWriteByte(b byte) error {
	if !tx.dw.inTx {
		return errNotInTx
	}
	avail := tx.dw.writeBuf[tx.off:]
	if len(avail) < 1 {
		return errShortWriteBuffer
	}
	avail[0] = b
	tx.off++
	return nil
}

// DoTxI2C executes the transaction over an I2C bus.
//
// This performs a combined write-read I2C transaction, first sending the
// register address and any data added to the transaction, then reading
// the specified number of bytes from the device.
//
// Parameters:
//   - bus: The I2C bus to communicate over
//   - deviceAddr: The I2C address of the target device
//   - readLength: Number of bytes to read from the device
//
// Returns the read data as a slice of the internal read buffer, valid until
// the next transaction. The transaction is automatically freed after execution.
//
// Returns an error if:
//   - No transaction is active (errNotInTx)
//   - The read buffer is too short (errShortReadBuffer)
//   - The I2C transaction fails
func (tx *Tx8) DoTxI2C(bus drivers.I2C, deviceAddr uint16, readLength int) ([]byte, error) {
	if tx.off == 0 || !tx.dw.inTx {
		return nil, errNotInTx
	}
	defer tx.freeTx()
	if len(tx.dw.readBuf) < readLength {
		return nil, errShortReadBuffer
	}
	rbuf := tx.dw.readBuf[:readLength]
	err := bus.Tx(deviceAddr, tx.dw.writeBuf[:tx.off], rbuf)
	if err != nil {
		return nil, err
	}
	return rbuf, err
}

// DoTxSPI executes the transaction over an SPI bus.
//
// This performs a full-duplex SPI transaction, simultaneously writing the
// register address and data while reading the same number of bytes from the device.
//
// If no read buffer was configured (readBuf is nil), this performs a write-only
// transaction and returns nil without error.
//
// Parameters:
//   - bus: The SPI bus to communicate over
//
// Returns the read data as a slice of the internal read buffer (same length as
// the write data), valid until the next transaction. The transaction is
// automatically freed after execution.
//
// Returns an error if:
//   - No transaction is active (errNotInTx)
//   - The read buffer is too short (errShortReadBuffer)
//   - The SPI transaction fails
func (tx *Tx8) DoTxSPI(bus drivers.SPI) (readBuf []byte, err error) {
	if tx.off == 0 || !tx.dw.inTx {
		return nil, errNotInTx
	}
	defer tx.freeTx()
	if tx.dw.readBuf == nil {
		err = bus.Tx(tx.dw.writeBuf[:tx.off], nil) // Special case, only use write buffer functionality.
		return nil, err
	} else if len(readBuf) < tx.off {
		return nil, errShortReadBuffer
	}
	rbuf := tx.dw.readBuf[:tx.off]
	err = bus.Tx(tx.dw.writeBuf[:tx.off], rbuf)
	if err != nil {
		return nil, err
	}
	return rbuf, err
}

// freeTx marks the transaction as complete, allowing a new transaction to be started.
// This is called internally by DoTxI2C and DoTxSPI after the transaction completes.
func (tx *Tx8) freeTx() {
	tx.dw.inTx = false
}
