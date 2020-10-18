package drivers

// SPI represents a SPI bus. It is implemented by the machine.SPI type.
type SPI interface {
	// Tx transmits the given buffer w and receives at the same time the buffer r.
	// The two buffers must be the same length. The only exception is when w or r are nil,
	// in which case Tx only transmits (without receiving) or only receives (while sending 0 bytes).
	Tx(w, r []byte) error

	// Transfer writes a single byte out on the SPI bus and receives a byte at the same time.
	// If you want to transfer multiple bytes, it is more efficient to use Tx instead.
	Transfer(b byte) (byte, error)
}
