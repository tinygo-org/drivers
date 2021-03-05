package drivers

import "io"

// UART represents a UART connection. It is implemented by the machine.UART
// type.
type UART interface {
	io.Reader
	io.Writer

	Buffered() int
}
