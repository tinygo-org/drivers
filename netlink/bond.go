package netlink

// Bond aggregates links as one
type Bond struct {
	*Bridge
}

func NewBond(links []Netlinker) *Bond {
	return &Bond{Bridge: NewBridge(links)}
}

func (b *Bond) SendEth(pkt []byte) error {
	// Send Ethernet pkt to active link(s)
	return nil
}
