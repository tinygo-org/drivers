package drivers

// I2C represents an I2C bus. It is notably implemented by the
// machine.I2C type.
type I2C interface {
	ReadRegister(addr uint8, r uint8, buf []byte) error
	WriteRegister(addr uint8, r uint8, buf []byte) error
	Tx(addr uint16, w, r []byte) error
}
