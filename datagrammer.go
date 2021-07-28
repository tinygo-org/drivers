package drivers

import (
	"io"
	"time"
)

// Datagrammer represents a reader/writer of data packets received over stream.
// These packets are low level representations of what could be an Ethernet/IP/TCP transaction.
// An example of an IC which implements this is the ENC28J60.
type Datagrammer interface {
	PacketWriter
	PacketReader
}

// Packet represents a handle to a packet in an underlying stream of data.
type Packet interface {
	io.Reader
	// Discard discards packet data. Reader is terminated as well.
	// If reader already terminated then it should have no effect.
	Discard() error
}

// PacketReader returns a handle to a packet. Ideally there should be no more
// than one active handle at a time.
type PacketReader interface {
	// Returns a Reader that reads from the next packet.
	NextPacket(deadline time.Time) (Packet, error)
}

// PacketWriter handles writes to buffer. Writes are not sent over stream until Flush is called.
type PacketWriter interface {
	io.Writer
	// Flush writes buffer to the underlying stream.
	Flush() error
}
