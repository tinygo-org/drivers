package tester

// MaxRegisters is the maximum number of registers supported for a Device.
const MaxRegisters = 255

type I2CDevice interface {
	// ReadRegister implements I2C.ReadRegister.
	readRegister(r uint8, buf []byte) error

	// WriteRegister implements I2C.WriteRegister.
	writeRegister(r uint8, buf []byte) error

	// Tx implements I2C.Tx
	Tx(w, r []byte) error

	// Addr returns the Device address.
	Addr() uint8
}
