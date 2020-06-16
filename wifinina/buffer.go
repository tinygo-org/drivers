package wifinina

import (
	"errors"
	"fmt"
	"io"
	"time"
)

// Buffer wraps a byte slice that is allocated for creating and/or parsing
// wifinina command/reply messages (which have the same format as each other).
// It is NOT safe for concurrent use.
//
// Cmd Struct Message
// ._______________________________________________________________________.
// | START CMD | C/R  . CMD  | N.PARAM | PARAM LEN | PARAM  | .. | END CMD |
// |___________|_____________|_________|___________|________|____|_________|
// |   8 bit   | 1bit . 7bit |  8bit   | 8/16* bit | nbytes | .. |   8bit  |
// |___________|_____________|_________|___________|________|____|_________|
// for most commands, the param len is 1 bytes; for data commands, it is 2
//
type Buffer struct {
	// buf is the internal storage for the buffer; its capacity is pre-determined
	// and allocated when the Buffer is created
	buf []byte
}

const (
	nParamsPos = 2

	defaultCapacity = 2048

	flagReply = 1 << 7
	flagData  = 0x40
)

var ErrBufferFull = errors.New("buffer is full")

// NewBuffer constructs a Buffer with the provided capacity. Enough memory to
// satisfy that capacity is allocated when the Buffer is constructed so that
// using the Buffer does not result in further allocation or garbage collection.
func NewBuffer(capacity int) *Buffer {
	return &Buffer{
		buf: make([]byte, 0, capacity),
	}
}

// StartCmd resets the buffer to be a request buffer, adds the StartCmd byte,
// the command/reply flag, the command byte, and sets the nparams byte to 0.
func (b *Buffer) StartCmd(cmd uint8) *Buffer {
	b.reset()
	b.append(CmdStart, cmd & ^(uint8(flagReply)), 0)
	return b
}

// AddData appends a byte slice as a data parameter to the Buffer.
func (b *Buffer) AddData(p []byte) {
	b.paramLen(len(p))
	b.append(p...) // copy the data to the internal buffer
	b.buf[nParamsPos]++
}

// AddString appends a string parameter to the Buffer.  The string should not
// be more than 256 bytes.
func (b *Buffer) AddString(s string) {
	// FIXME: based on protocol strings over 256 bytes should not possible, check?
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

// EndCmd denotes that the last param has been added to buffer; this will
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
	return (b.buf[1] & 0x70) == flagData
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
// the length of a parameter. For most commands, 1; for data commands, can be 2
func (b *Buffer) ParamLenSize() uint8 {
	if b.IsDataCommand() {
		cmd, rep := b.Command(), b.IsReply()
		if (cmd == CmdGetDatabufTCP) || (!rep && cmd != CmdGetDatabufTCP) {
			return 2
		}
	}
	return 1
}

// paramLen adds the length of the next parameter to the buffer
func (b *Buffer) paramLen(l int) {
	if b.ParamLenSize() == 2 {
		b.append(byte(l>>8), byte(l&0xFF))
	} else {
		b.append(byte(l))
	}
}

// Len returns the length of the buffer in bytes
func (b *Buffer) Len() int {
	return len(b.buf)
}

func (b *Buffer) WriteTo(w io.Writer) (int, error) {
	return w.Write(b.buf)
}

func (b *Buffer) Bytes() []byte {
	return b.buf
}

type ByteTransferer interface {
	Transfer(b byte) (byte, error)
}

// ParseReply switches the buffer from being a request buffer to a reply buffer
// and attempts to parse the entire reply from the given byte source.  If the
// checkCmd argument is a non-zero value, it will be compared to the command
// byte in the reply and if they do not match ErrIncorrectReply is returned.
func (b *Buffer) ReadReply(r ByteTransferer, checkCmd byte) (err error) {
	b.reset()

	// first we will loop until we either get CmdStart or CmdErr from the reader
	// which would suggest that the reply is starting
	// TODO: should this timeout be parameterized or configurable?
	// TODO: does any extra info come with CmdErr?
	var read byte
	for now := time.Now(); time.Since(now) < 5*time.Millisecond; {
		if read, err = r.Transfer(0xFF); err != nil {
			return err
		}
		if read == CmdErr {
			return ErrCmdErrorReceived
		}
		if read == CmdStart {
			b.add(read)
			break
		}
	}
	if read != CmdStart {
		return ErrCheckStartCmd
	}

	// check to ensure next byte is the command byte with the reply flag set
	if read, err = r.Transfer(0xFF); err != nil {
		return err
	}
	if checkCmd != 0 {
		if read != (flagReply | checkCmd) {
			return ErrIncorrectReply
		}
	}
	b.add(read)

	// next byte should be the number of parameters
	if read, err = r.Transfer(0xFF); err != nil {
		return err
	}
	b.add(read)

	// depending on whether or not this is a data command, the parameter length
	// will either be 8 bits or 16 bits, respectively
	var readParamLen func(r ByteTransferer) (uint16, error)
	if b.Command() == CmdGetDatabufTCP {
		readParamLen = b.read16
	} else {
		readParamLen = b.read8
	}

	// loop over the parameters and read them into the buffer
	for i, numParams, pLen := 0, int(read), uint16(0); i < numParams; i++ {
		// read in the parameter based in the given parameter length
		pLen, err = readParamLen(r)
		if err != nil {
			return err
		}
		for j := uint16(0); j < pLen; j++ {
			if read, err = r.Transfer(0xFF); err != nil {
				return err
			}
			b.add(read)
		}
	}

	// check to ensure next byte is the end command byte
	if read, err = r.Transfer(0xFF); err != nil {
		return err
	}
	b.add(read)
	if read != CmdEnd {
		return ErrIncorrectSentinel
	}

	return nil
}

func (b *Buffer) GetByteParam(n int, v *byte) error {
	if sl, err := b.paramSlice(n); err != nil {
		return err
	} else {
		if len(sl) != 1 {
			/*
				if _debug {
					println("expected length 1, was actually", len(sl), "\r")
					_ = PrintBuffer(b, os.Stdout)
				}
			*/
			return ErrUnexpectedLength
		}
		*v = sl[0]
		return nil
	}
}

func (b *Buffer) GetUint16Param(n int, v *uint16) error {
	if sl, err := b.paramSlice(n); err != nil {
		return err
	} else {
		if len(sl) != 2 {
			/*
				if _debug {
					println("expected length 2, was actually", len(sl), "\r")
					_ = PrintBuffer(b, os.Stdout)
				}
			*/
			return ErrUnexpectedLength
		}
		*v = (uint16(sl[1]) << 8) | (uint16(sl[0]))
		return nil
	}
}

func (b *Buffer) GetUint32Param(n int, v *uint32) error {
	if sl, err := b.paramSlice(n); err != nil {
		return err
	} else {
		if len(sl) != 4 {
			/*
				if _debug {
					println("expected length 4, was actually", len(sl), "\r")
					_ = PrintBuffer(b, os.Stdout)
				}
			*/
			return ErrUnexpectedLength
		}
		*v = (uint32(sl[3]) << 24) | (uint32(sl[2]) << 16) |
			(uint32(sl[1]) << 8) | (uint32(sl[0]))
		return nil
	}
}

func (b *Buffer) GetUint64Param(n int, v *uint64) error {
	if sl, err := b.paramSlice(n); err != nil {
		return err
	} else {
		if len(sl) != 6 {
			/*
				if _debug {
					println("expected length 6, was actually", len(sl), "\r")
				}
			*/
			return ErrUnexpectedLength
		}
		*v = (uint64(sl[5]) << 56) | (uint64(sl[4]) << 48) | (uint64(sl[3]) << 40) |
			(uint64(sl[2]) << 32) | (uint64(sl[1]) << 24) | (uint64(sl[0]) << 16)
		return nil
	}
}

func (b *Buffer) GetStringParam(n int, v *string) error {
	if sl, err := b.paramSlice(n); err != nil {
		return err
	} else {
		*v = string(sl)
		return nil
	}
}

func (b *Buffer) GetBufferParam(n int, v []byte) (int, error) {
	if sl, err := b.paramSlice(n); err != nil {
		return 0, err
	} else {
		return copy(v, sl), nil
	}
}

func (b *Buffer) paramSlice(n int) ([]byte, error) {
	if np := int(b.NumParams()); n >= np {
		return nil, fmt.Errorf("n %d is greater than number of params %d", n, np)
	}
	var pLenSize = int(b.ParamLenSize())
	var pos = 3
	var sl = b.buf
	for i := 0; i < int(b.NumParams()); i++ {
		pLen := int(sl[pos])
		if pLenSize == 2 {
			pLen = (pLen << 8) | int(sl[pos+1])
		}
		if i == n {
			return sl[pos+pLenSize : pos+pLenSize+pLen], nil
		}
		pos += pLenSize + pLen
	}
	return nil, ErrIncorrectReply
}

// reads 2 bytes, adds them to the buffer, and returns them as a uint16
// this function is used in ReadReply to read 16-bit parameter length values
func (b *Buffer) read16(r ByteTransferer) (v uint16, err error) {
	var bite byte
	if bite, err = r.Transfer(0xFF); err == nil {
		v |= uint16(bite) << 8
		b.add(bite)
		if bite, err = r.Transfer(0xFF); err == nil {
			v |= uint16(bite)
			b.add(bite)
		}
	}
	return
}

// reads 1 byte, adds it to the buffer, and returns it as a uint16
// this function is used in ReadReply to read 8-bit parameter length values
func (b *Buffer) read8(r ByteTransferer) (v uint16, err error) {
	var bite byte
	if bite, err = r.Transfer(0xFF); err == nil {
		v = uint16(bite)
		b.add(bite)
	}
	return
}

// resets the buffer's internal storage making it ready to buffer a new command
func (b *Buffer) reset() {
	b.buf = b.buf[:0]
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
