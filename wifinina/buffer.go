package wifinina

import (
	"errors"
	"io"
)

// Cmd Struct Message */
// ._______________________________________________________________________.
// | START CMD | C/R  . CMD  | N.PARAM | PARAM LEN | PARAM  | .. | END CMD |
// |___________|_____________|_________|___________|________|____|_________|
// |   8 bit   | 1bit . 7bit |  8bit   | 8/16* bit | nbytes | .. |   8bit  |
// |___________|_____________|_________|___________|________|____|_________|
// * for most commands, the param len is 1 bytes; for data commands, it is 2

const (
	nParamsPos = 2

	defaultCapacity = 4096
)

type Buffer struct {
	//bytes.Buffer
	// buf is the internal storage for the buffer; its capacity is pre-determined
	// and allocated when the Buffer is created
	buf []byte // contents are the bytes buf[off : len(buf)]
	off int    // read at &buf[off], write at &buf[len(buf)]
}

var ErrBufferFull = errors.New("buffer is full")

// NewBuffer constructs a Buffer with the provided capacity. Enough memory to
// satisfy that capacity is allocated when the Buffer is constructed so that
// using the Buffer does not result in further allocation or garbage collection.
func NewBuffer(capacity int) *Buffer {
	return &Buffer{
		buf: make([]byte, 0, capacity),
	}
}

// StartCmd resets the buffer to be a request buffer, adds the StartCmd, the
// command/reply flag, and the command byte, and sets the nparams byte to 0.
func (b *Buffer) StartCmd(cmd uint8) *Buffer {
	b.reset()
	b.append(CmdStart, cmd & ^(uint8(FlagReply)), 0)
	return b
}

// AddData appends a byte slice as a data parameter to the Buffer.
func (b *Buffer) AddData(p []byte) {
	// note: it appears that a data buffer is the only wifinina parameter that
	// has a two-byte length
	// TODO: consider adding a scheme to keep track of data parameters
	//l := len(p)
	//b.append(byte(l>>8), byte(l&0xFF)) // write the parameter length
	b.paramLen(len(p))
	b.append(p...) // copy the data to the internal buffer
	b.buf[nParamsPos]++
}

// AddString appends a string parameter to the Buffer.  The string should not
// be more than 256 bytes.
func (b *Buffer) AddString(s string) {
	// FIXME: based on protocol strings over 256 bytes should not possible, check?
	//b.add(byte(len(s)))
	b.paramLen(len(s))
	b.append([]byte(s)...)
	b.buf[nParamsPos]++
}

// AddByte appends a byte parameter to the Buffer.
func (b *Buffer) AddByte(p byte) {
	b.paramLen(1)
	b.add(p)
	b.buf[nParamsPos]++
}

// AddUint16 appends a uint16 parameter to the buffer.
func (b *Buffer) AddUint16(p uint16) {
	b.paramLen(2)
	b.append(uint8(p>>8), uint8(p&0xFF))
	b.buf[nParamsPos]++
}

// AddUint32 appends a uint32 parameter to the buffer.
func (b *Buffer) AddUint32(p uint32) {
	b.paramLen(4)
	b.append(uint8(p>>24), uint8(p>>16), uint8(p>>8), uint8(p&0xFF))
	b.buf[nParamsPos]++
}

// EndCmd denotes that the last param has been added to buffer; this will set
// append the CmdEnd byte and adds necessary padding
func (b *Buffer) EndCmd() {
	padding := (4 - ((len(b.buf) + 1) % 4)) & 3
	b.append([]byte{CmdEnd, 0xFF, 0xFF, 0xFF, 0xFF}[:padding+1]...)
}

// Command returns the command byte that is set in the buffer
func (b *Buffer) Command() byte {
	return b.buf[1] & 0x7f
}

// IsDataCommand checks if this is a data command that takes buffers as params
func (b *Buffer) IsDataCommand() bool {
	return (b.buf[1] & 0x70) == 0x40
}

// IsReply returns whether or not the reply flag is set
func (b *Buffer) IsReply() bool {
	return (b.buf[1] & 0x80) > 0
}

// NumParams returns the number of parameters that have been set on the command
func (b *Buffer) NumParams() uint8 {
	return b.buf[2]
}

// ParamLenSize returns the number of bytes used in a wifinina message to send
// the length of a parameter. For most commands, 1; for data commands, 2
func (b *Buffer) ParamLenSize() uint8 {
	if b.IsDataCommand() {
		return 2
	} else {
		return 1
	}
}

func (b *Buffer) paramLen(l int) {
	if b.IsDataCommand() {
		b.append(byte(l>>8), byte(l&0xFF))
	} else {
		b.append(byte(l))
	}
}

func (b *Buffer) Len() int {
	return len(b.buf)
}

func (b *Buffer) WriteTo(w io.Writer) (int, error) {
	return w.Write(b.buf)
}

func (b *Buffer) Bytes() []byte {
	return b.buf
}

// resets the buffer's internal storage making it ready to buffer a new command
func (b *Buffer) reset() {
	b.buf = b.buf[:0]
	b.off = 0
	//b.err = nil
}

// (copied from bytes.Buffer)
// tryGrowByReslice is a inlineable version of grow for the fast-case where the
// internal buffer only needs to be resliced.
// It returns the index where bytes should be written and whether it succeeded.
func (b *Buffer) tryGrowByReslice(n int) (int, bool) {
	if b.buf == nil {
		b.buf = make([]byte, 0, defaultCapacity)
	}
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}

// appends bytes the buffer and panics with ErrBufferFull if no more bytes can
// be written.
func (b *Buffer) append(c ...byte) {
	m, ok := b.tryGrowByReslice(len(c))
	if !ok {
		panic(ErrBufferFull)
	}
	copy(b.buf[m:], c)
}

// adds a single byte the buffer and panics with ErrBufferFull if no more bytes
// can be written.
func (b *Buffer) add(c byte) {
	m, ok := b.tryGrowByReslice(1)
	if !ok {
		panic(ErrBufferFull)
	}
	b.buf[m] = c
}
