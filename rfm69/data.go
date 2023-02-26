package rfm69

// Data is the data structure for the protocol
type Data struct {
	ToAddress   byte
	FromAddress byte
	Data        []byte
	RequestAck  bool
	SendAck     bool
	Rssi        int
}

// ToAck creates an ack
func (d *Data) ToAck() *Data {
	return &Data{
		ToAddress: d.FromAddress,
		SendAck:   true,
	}
}
