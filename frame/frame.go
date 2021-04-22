package frame

type Framer interface {

	// FrameLength includes header and data length in bytes
	FrameLength() uint16
	// MarshalFrame writes frame and subframes into buffer and returns number of bytes written
	MarshalFrame([]byte) (uint16, error)
	UnmarshalFrame([]byte) error
	// Clear Options removes optional fields which would otherwise give an artificially high
	// FrameLength when Unmarshalling
	SetResponse() error
}
