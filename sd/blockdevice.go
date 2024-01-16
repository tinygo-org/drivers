package sd

import (
	"errors"
	"math/bits"
)

var (
	errNegativeOffset = errors.New("sd: negative offset")
)

// Compile time guarantee of interface implementation.
var _ Card = (*SPICard)(nil)

type Card interface {
	// WriteBlocks writes the given data to the card, starting at the given block index.
	// The data must be a multiple of the block size.
	WriteBlocks(data []byte, startBlockIdx int64) error
	// ReadBlocks reads the given number of blocks from the card, starting at the given block index.
	// The dst buffer must be a multiple of the block size.
	ReadBlocks(dst []byte, startBlockIdx int64) error
	// EraseBlocks erases
	EraseSectors(startBlockIdx, numBlocks int64) error
}

// NewBlockDevice creates a new BlockDevice from a Card.
func NewBlockDevice(card Card, blockSize int, numBlocks, eraseBlockSizeInBytes int64) (*BlockDevice, error) {
	if card == nil || blockSize <= 0 || eraseBlockSizeInBytes <= 0 || numBlocks <= 0 {
		return nil, errors.New("invalid argument(s)")
	}
	tz := bits.TrailingZeros(uint(blockSize))
	if blockSize>>tz != 1 {
		return nil, errors.New("blockSize must be a power of 2")
	}
	bd := &BlockDevice{
		card:           card,
		blockbuf:       make([]byte, blockSize),
		blockshift:     tz,
		blockmask:      (1 << tz) - 1,
		numblocks:      int64(numBlocks),
		eraseBlockSize: eraseBlockSizeInBytes,
	}
	return bd, nil
}

// BlockDevice implements tinyfs.BlockDevice interface for an [sd.Card] type.
type BlockDevice struct {
	card           Card
	blockbuf       []byte
	blockshift     int
	blockmask      int64
	numblocks      int64
	eraseBlockSize int64
}

func (bd *BlockDevice) moduloBlockSize(n int64) int64 {
	return n & bd.blockmask
}

func (bd *BlockDevice) divideBlockSize(n int64) int64 {
	return n >> bd.blockshift
}

// ReadAt implements [io.ReadAt] interface for an SD card.
func (bd *BlockDevice) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errNegativeOffset
	}
	blockSize := len(bd.blockbuf)
	blockIdx := bd.divideBlockSize(off)
	blockOff := bd.moduloBlockSize(off)
	if blockOff != 0 {
		// Non-aligned first block case.
		if err := bd.card.ReadBlocks(bd.blockbuf, blockIdx); err != nil {
			return n, err
		}
		n += copy(p, bd.blockbuf[blockOff:])
		p = p[n:]
		blockIdx++
	}

	remaining := len(p) - n
	if remaining >= blockSize {
		// 1 or more full blocks case.
		endOffset := remaining - int(bd.moduloBlockSize(int64(remaining)))
		err = bd.card.ReadBlocks(p[:endOffset], blockIdx)
		if err != nil {
			return n, err
		}
		p = p[endOffset:]
		n += endOffset
		blockIdx += int64(endOffset / blockSize)
	}

	if len(p) > 0 {
		// Non-aligned last block case.
		if err := bd.card.ReadBlocks(bd.blockbuf, blockIdx); err != nil {
			return n, err
		}
		n += copy(p, bd.blockbuf)
	}
	return n, nil
}

// WriteAt implements [io.WriterAt] interface for an SD card.
func (bd *BlockDevice) WriteAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errNegativeOffset
	}
	blockSize := len(bd.blockbuf)
	blockIdx := bd.divideBlockSize(off)
	blockOff := bd.moduloBlockSize(off)
	if blockOff != 0 {
		// Non-aligned first block case.
		if err := bd.card.ReadBlocks(bd.blockbuf, blockIdx); err != nil {
			return n, err
		}
		n += copy(bd.blockbuf[blockOff:], p)
		if err := bd.card.WriteBlocks(bd.blockbuf, blockIdx); err != nil {
			return n, err
		}
		p = p[n:]
		blockIdx++
	}

	remaining := len(p) - n
	if remaining >= blockSize {
		// 1 or more full blocks case.
		endOffset := remaining - int(bd.moduloBlockSize(int64(remaining)))
		err = bd.card.WriteBlocks(p[:endOffset], blockIdx)
		if err != nil {
			return n, err
		}
		p = p[endOffset:]
		n += endOffset
		blockIdx += int64(endOffset / blockSize)
	}

	if len(p) > 0 {
		// Non-aligned last block case.
		if err := bd.card.ReadBlocks(bd.blockbuf, blockIdx); err != nil {
			return n, err
		}
		n += copy(bd.blockbuf, p)
		if err := bd.card.WriteBlocks(bd.blockbuf, blockIdx); err != nil {
			return n, err
		}
	}
	return n, nil
}

// Size returns the number of bytes in this block device.
func (bd *BlockDevice) Size() int64 {
	return int64(len(bd.blockbuf)) * bd.numblocks
}

// EraseBlocks erases the given number of blocks. An implementation may
// transparently coalesce ranges of blocks into larger bundles if the chip
// supports this. The start and len parameters are in block numbers, use
// EraseBlockSize to map addresses to blocks.
func (bd *BlockDevice) EraseBlocks(startEraseBlockIdx, len int64) error {
	return bd.card.EraseSectors(startEraseBlockIdx, len)
}

// EraseBlockSize returns the smallest erasable area on this particular chip
// in bytes. This is used for the block size in EraseBlocks.
// It must be a power of two, and may be as small as 1. A typical size is 4096.
func (bd *BlockDevice) EraseBlockSize() int64 {
	return bd.eraseBlockSize
}
