package bno08x

import (
	"encoding/binary"
	"time"
)

// halI2C implements the hardware abstraction layer for I2C communication.
type halI2C struct {
	device    *Device
	chunkSize int
	scratch   []byte
	header    [shtpHeaderLength]byte // Reusable header buffer
}

func newHAL(dev *Device) *halI2C {
	chunk := dev.readChunk
	if chunk < shtpHeaderLength {
		chunk = shtpHeaderLength
	}
	return &halI2C{
		device:    dev,
		chunkSize: chunk,
		scratch:   make([]byte, chunk),
	}
}

func (h *halI2C) open() error {
	// HAL is now open and ready for communication
	// Soft reset will be sent after handlers are registered
	return nil
}

func (h *halI2C) close() {}

func (h *halI2C) read(target []byte) (int, uint32, error) {
	// Read SHTP header (4 bytes) to get packet length
	// Use pre-allocated header buffer to avoid allocations
	err := h.device.bus.Tx(h.device.address, nil, h.header[:])
	if err != nil {
		return 0, 0, err
	}

	// Parse packet length from header
	packetLen := binary.LittleEndian.Uint16(h.header[0:2])

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
			request = h.chunkSize
			if request > cargoRemaining {
				request = cargoRemaining
			}
		} else {
			// Subsequent reads: each chunk has a 4-byte header we need to skip
			request = h.chunkSize
			if request > cargoRemaining+shtpHeaderLength {
				request = cargoRemaining + shtpHeaderLength
			}
		}

		// Ensure scratch buffer is large enough
		if request > len(h.scratch) {
			h.scratch = make([]byte, request)
		}
		buf := h.scratch[:request]

		// Read chunk
		err = h.device.bus.Tx(h.device.address, nil, buf)
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

	timestamp := uint32(time.Now().UnixNano() / 1000)
	return int(packetLen), timestamp, nil
}

func (h *halI2C) write(frame []byte) (int, error) {
	if len(frame) > h.chunkSize {
		return 0, errFrameTooLarge
	}
	err := h.device.bus.Tx(h.device.address, frame, nil)
	if err != nil {
		return 0, err
	}
	return len(frame), nil
}

func (h *halI2C) getTimeUs() uint32 {
	return uint32(time.Now().UnixNano() / 1000)
}
