package frame

type Framer interface {
	// FrameLength includes header and data length in bytes
	FrameLength() uint16
	MarshalFrame([]byte) error
	UnmarshalFrame([]byte) error
	// Clear Options removes optional fields which would otherwise give an artificially high
	// FrameLength when Unmarshalling
	ClearOptions()
}
