package bno08x

import (
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/internal/pin"
)

// I2CConfig holds I2C-specific configuration options.
type I2CConfig struct {
	// Address is the I2C address (default: 0x4A).
	Address uint16

	// ResetPin is the optional hardware reset pin.
	ResetPin pin.OutputFunc

	// ReadChunk is the I2C read chunk size (default: 32 bytes).
	ReadChunk int
}

const (
	// DefaultAddress is the default I2C address.
	DefaultAddress = 0x4A
)

// NewI2C creates a new BNO08x device using I2C communication.
func NewI2C(bus drivers.I2C) *Device {
	return &Device{
		bus: &I2CBus{
			wire:      bus,
			address:   DefaultAddress,
			readChunk: i2cDefaultChunk,
		},
	}
}

// I2CBus implements the Buser interface for I2C communication.
type I2CBus struct {
	wire      drivers.I2C
	address   uint16
	readChunk int
	scratch   []byte
	header    [shtpHeaderLength]byte
}

// configure sets up the I2C bus with the specified address and chunk size.
func (b *I2CBus) configure(address uint16, readChunk int) error {
	if address != 0 {
		b.address = address
	}
	if readChunk > 0 {
		b.readChunk = readChunk
	}

	chunk := b.readChunk
	if chunk < shtpHeaderLength {
		chunk = shtpHeaderLength
	}
	b.scratch = make([]byte, chunk)

	return nil
}

// read reads data from the I2C bus.
func (b *I2CBus) read(target []byte) (int, uint32, error) {
	// Read SHTP header (4 bytes) to get packet length
	// Use pre-allocated header buffer to avoid allocations
	err := b.wire.Tx(b.address, nil, b.header[:])
	if err != nil {
		return 0, 0, err
	}

	// Parse packet length from header
	packetLen := uint16(b.header[0]) | (uint16(b.header[1]) << 8)

	// Check if continuation bit is set (0x8000)
	// This means no data is available yet
	if packetLen&continueMask != 0 {
		return 0, 0, nil
	}

	// No continuation bit, check for actual data
	if packetLen == 0 {
		return 0, 0, nil
	}

	if int(packetLen) > len(target) {
		return 0, 0, errBufferTooSmall
	}

	// Now read the full packet in chunks, re-reading the header in first chunk
	// This follows Arduino's approach: initial header read is just to get size,
	// actual packet data (including header) is read in the loop
	cargoRemaining := int(packetLen)
	offset := 0
	firstRead := true

	for cargoRemaining > 0 {
		var request int
		if firstRead {
			// First read: get the full packet including header (up to chunkSize)
			request = b.readChunk
			if request > cargoRemaining {
				request = cargoRemaining
			}
		} else {
			// Subsequent reads: each chunk has a 4-byte header we need to skip
			request = b.readChunk
			if request > cargoRemaining+shtpHeaderLength {
				request = cargoRemaining + shtpHeaderLength
			}
		}

		// Ensure scratch buffer is large enough
		if request > len(b.scratch) {
			b.scratch = make([]byte, request)
		}
		buf := b.scratch[:request]

		// Read chunk
		err = b.wire.Tx(b.address, nil, buf)
		if err != nil {
			return 0, 0, err
		}

		var cargoRead int
		if firstRead {
			// First read: copy everything including header
			cargoRead = request
			copy(target[offset:], buf[:cargoRead])
			firstRead = false
		} else {
			// Subsequent reads: skip the 4-byte header
			cargoRead = request - shtpHeaderLength
			copy(target[offset:], buf[shtpHeaderLength:shtpHeaderLength+cargoRead])
		}

		offset += cargoRead
		cargoRemaining -= cargoRead
	}

	// Extract timestamp from the header in the target buffer
	timestamp := uint32(target[2]) | (uint32(target[3]) << 8)

	return int(packetLen), timestamp, nil
}

// write sends data over the I2C bus.
func (b *I2CBus) write(data []byte) error {
	return b.wire.Tx(b.address, data, nil)
}

// softReset sends a soft reset command via I2C.
func (b *I2CBus) softReset() error {
	// Send soft reset packet via I2C as per Adafruit implementation
	// Format: [length_low, length_high, channel, sequence, command]
	// This is: 5 bytes total, channel 1 (executable), command 1 (reset)
	softResetPacket := []byte{5, 0, 1, 0, 1}

	// Try up to 5 times
	var err error
	for i := 0; i < 5; i++ {
		err = b.wire.Tx(b.address, softResetPacket, nil)
		if err == nil {
			// Success - wait for sensor to process reset
			time.Sleep(300 * time.Millisecond)
			return nil
		}
		time.Sleep(30 * time.Millisecond)
	}
	return err
}
