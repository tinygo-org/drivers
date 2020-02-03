package flash

import "fmt"

type JedecID struct {
	ManufID  uint8
	MemType  uint8
	Capacity uint8
}

func (id *JedecID) AsUint32() uint32 {
	return uint32(id.ManufID)<<16 | uint32(id.MemType)<<8 | uint32(id.Capacity)
}

func (id JedecID) String() string {
	return fmt.Sprintf("%06X", id.AsUint32())
}

type Attrs struct {
	TotalSize uint32
	//start_up_time_us uint16

	// Three response bytes to 0x9f JEDEC ID command.
	JedecID

	// Max clock speed for all operations and the fastest read mode.
	MaxClockSpeedMHz uint8

	// Bitmask for Quad Enable bit if present. 0x00 otherwise. This is for the
	// highest byte in the status register.
	QuadEnableBitMask uint8

	HasSectorProtection bool

	// Supports the 0x0b fast read command with 8 dummy cycles.
	SupportsFastRead bool

	// Supports the fast read, quad output command 0x6b with 8 dummy cycles.
	SupportsQSPI bool

	// Supports the quad input page program command 0x32. This is known as 1-1-4
	// because it only uses all four lines for data.
	SupportsQSPIWrites bool

	// Requires a separate command 0x31 to write to the second byte of the status
	// register. Otherwise two byte are written via 0x01.
	WriteStatusRegisterSplit bool

	// True when the status register is a single byte. This implies the Quad
	// Enable bit is in the first byte and the Read Status Register 2 command
	// (0x35) is unsupported.
	SingleStatusByte bool
}
