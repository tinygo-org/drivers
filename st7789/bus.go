package st7789

// Bus is the interface that wraps the basic Tx and Transfer methods
// for communication buses like SPI or Parallel.
type Bus interface {
	Tx(w, r []byte) error
	Transfer(w byte) (byte, error)
}
