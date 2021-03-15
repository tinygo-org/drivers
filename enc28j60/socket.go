package enc28j60

// Socket w5500
type Socket struct {
	d *Dev

	Num uint8

	receivedSize  uint16
	receiveOffset uint16
}
