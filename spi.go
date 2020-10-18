package drivers

// SPI represents a SPI bus. It is implemented by the machine.SPI type.
type SPI interface {
	Tx(w, r []byte) error
	Transfer(b byte) (byte, error)
}
