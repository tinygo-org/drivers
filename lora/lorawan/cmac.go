package lorawan

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"hash"
	"unsafe"
)

type cmacHash struct {
	ciph cipher.Block
	k1   []byte
	k2   []byte
	data []byte
	x    []byte
}

const (
	Size      = aes.BlockSize
	blockSize = Size
)

var (
	subkeyZero []byte
	subkeyRb   []byte
)

func init() {
	subkeyZero = bytes.Repeat([]byte{0x00}, blockSize)
	subkeyRb = append(bytes.Repeat([]byte{0x00}, blockSize-1), 0x87)
}

// New returns an AES-CMAC hash using the supplied key. The key must be 16, 24,
// or 32 bytes long.
func NewCmac(key []byte) (hash.Hash, error) {
	// Create a cipher.
	ciph, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// Set up the hash object.
	h := &cmacHash{ciph: ciph}
	h.k1, h.k2 = generateSubkeys(ciph)
	h.Reset()
	return h, nil
}

func Xor(dst []byte, a []byte, b []byte) error {
	if len(dst) != len(a) || len(a) != len(b) {
		panic("crypto/Xor: bad length")
	}
	for i, _ := range a {
		dst[i] = a[i] ^ b[i]
	}
	return nil
}

func (h *cmacHash) Reset() {
	h.data = h.data[:0]
	h.x = make([]byte, blockSize)
}

func (h *cmacHash) BlockSize() int {
	return h.ciph.BlockSize()
}

func (h *cmacHash) Size() int {
	return h.ciph.BlockSize()
}

func (h *cmacHash) Sum(b []byte) []byte {
	dataLen := len(h.data)

	// We should have at most one block left.
	if dataLen > blockSize {
		panic("cmacHash err1")
	}

	// Calculate M_last.
	mLast := make([]byte, blockSize)
	if dataLen == blockSize {
		Xor(mLast, h.data, h.k1)
	} else {
		// TODO(jacobsa): Accept a destination buffer in common.PadBlock and
		// simplify this code.
		Xor(mLast, PadBlock(h.data), h.k2)
	}

	y := make([]byte, blockSize)
	Xor(y, mLast, h.x)

	result := make([]byte, blockSize)
	h.ciph.Encrypt(result, y)

	b = append(b, result...)
	return b
}

func (h *cmacHash) Write(p []byte) (n int, err error) {
	n = len(p)

	// First step: consume enough data to expand h.data to a full block, if
	// possible.
	{
		toConsume := blockSize - len(h.data)
		if toConsume > len(p) {
			toConsume = len(p)
		}

		h.data = append(h.data, p[:toConsume]...)
		p = p[toConsume:]
	}

	// If there's no data left in p, it means h.data might not be a full block.
	// Even if it is, we're not sure it's the final block, which we must treat
	// specially. So we must stop here.
	if len(p) == 0 {
		return
	}

	// h.data is a full block and is not the last; process it.
	h.writeBlocks(h.data)
	h.data = h.data[:0]

	// Consume any further full blocks in p that we're sure aren't the last. Note
	// that we're sure that len(p) is greater than zero here.
	blocksToProcess := (len(p) - 1) / blockSize
	bytesToProcess := blocksToProcess * blockSize

	h.writeBlocks(p[:bytesToProcess])
	p = p[bytesToProcess:]

	// Store the rest for later.
	h.data = append(h.data, p...)

	return
}

func (h *cmacHash) writeBlocks(p []byte) {
	y := make([]byte, blockSize)

	for off := 0; off < len(p); off += blockSize {
		block := p[off : off+blockSize]

		xorBlock(
			unsafe.Pointer(&y[0]),
			unsafe.Pointer(&h.x[0]),
			unsafe.Pointer(&block[0]))

		h.ciph.Encrypt(h.x, y)
	}

	return
}

func xorBlock(
	dstPtr unsafe.Pointer,
	aPtr unsafe.Pointer,
	bPtr unsafe.Pointer) {
	// Check assumptions. (These are compile-time constants, so this should
	// compile out.)
	const wordSize = unsafe.Sizeof(uintptr(0))
	if blockSize != 4*wordSize {
		panic("xorBlock err1")
	}

	// Convert.
	a := (*[4]uintptr)(aPtr)
	b := (*[4]uintptr)(bPtr)
	dst := (*[4]uintptr)(dstPtr)

	// Compute.
	dst[0] = a[0] ^ b[0]
	dst[1] = a[1] ^ b[1]
	dst[2] = a[2] ^ b[2]
	dst[3] = a[3] ^ b[3]
}

func PadBlock(block []byte) []byte {
	blockLen := len(block)
	if blockLen >= aes.BlockSize {
		panic("PadBlock input must be less than 16 bytes.")
	}

	result := make([]byte, aes.BlockSize)
	copy(result, block)
	result[blockLen] = 0x80

	return result
}

// Given the supplied cipher, whose block size must be 16 bytes, return two
// subkeys that can be used in MAC generation. See section 5.3 of NIST SP
// 800-38B. Note that the other NIST-approved block size of 8 bytes is not
// supported by this function.
func generateSubkeys(ciph cipher.Block) (k1 []byte, k2 []byte) {
	if ciph.BlockSize() != blockSize {
		panic("generateSubkeys requires a cipher with a block size of 16 bytes.")
	}

	// Step 1
	l := make([]byte, blockSize)
	ciph.Encrypt(l, subkeyZero)

	// Step 2: Derive the first subkey.
	if Msb(l) == 0 {
		// TODO(jacobsa): Accept a destination buffer in ShiftLeft and then hoist
		// the allocation in the else branch below.
		k1 = ShiftLeft(l)
	} else {
		k1 = make([]byte, blockSize)
		Xor(k1, ShiftLeft(l), subkeyRb)
	}

	// Step 3: Derive the second subkey.
	if Msb(k1) == 0 {
		k2 = ShiftLeft(k1)
	} else {
		k2 = make([]byte, blockSize)
		Xor(k2, ShiftLeft(k1), subkeyRb)
	}

	return
}

func ShiftLeft(b []byte) []byte {
	l := len(b)
	if l == 0 {
		panic("shiftLeft requires a non-empty buffer.")
	}

	output := make([]byte, l)

	overflow := byte(0)
	for i := int(l - 1); i >= 0; i-- {
		output[i] = b[i] << 1
		output[i] |= overflow
		overflow = (b[i] & 0x80) >> 7
	}

	return output
}

// Msb returns the most significant bit of the supplied data (which must be
// non-empty). This is the MSB(L) function of RFC 4493.
func Msb(buf []byte) uint8 {
	if len(buf) == 0 {
		panic("msb requires non-empty buffer.")
	}

	return buf[0] >> 7
}
