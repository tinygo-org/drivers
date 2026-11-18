package machine

import "io"

type UART struct {
	Reader io.Reader
	Writer io.Writer
}

func (u *UART) Read(b []byte) (int, error)  { return u.Reader.Read(b) }
func (u *UART) Write(b []byte) (int, error) { return u.Writer.Write(b) }
func (u *UART) Configure(UARTConfig) error  { return nil }
func (u *UART) Buffered() int               { return 0 }

type UARTConfig struct {
	TX, RX   Pin
	BaudRate uint32
}
type Pin uint8
type PinConfig struct{ Mode uint8 }

const PinOutput = 1

func (Pin) Configure(PinConfig) {}
func (Pin) Low()                {}
func (Pin) High()               {}
