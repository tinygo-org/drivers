package drivers

// I2C represents an I2C bus. It is notably implemented by the
// machine.I2C type.
type I2C interface {
	// Tx performs a [I²C] transaction with address addr.
	// Most I2C peripherals have some sort of register mapping scheme to allow
	// users to interact with them:
	//
	//  bus.Tx(addr, []byte{reg}, buf) // Reads register reg into buf.
	//  bus.Tx(addr, append([]byte{reg}, buf...), nil) // Writes buf into register reg.
	//
	// The semantics of most I2C transactions require that the w write buffer be non-empty.
	//
	// [I²C]: https://en.wikipedia.org/wiki/I%C2%B2C
	Tx(addr uint16, w, r []byte) error
}
