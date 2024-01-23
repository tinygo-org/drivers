package sd

import (
	"errors"
	"io"
	"math/bits"
)

var (
	errNegativeOffset = errors.New("sd: negative offset")
)

// Compile time guarantee of interface implementation.
var _ Card = (*SPICard)(nil)
var _ io.ReaderAt = (*BlockDevice)(nil)
var _ io.WriterAt = (*BlockDevice)(nil)

type Card interface {
	// WriteBlocks writes the given data to the card, starting at the given block index.
	// The data must be a multiple of the block size.
	WriteBlocks(data []byte, startBlockIdx int64) (int, error)
	// ReadBlocks reads the given number of blocks from the card, starting at the given block index.
	// The dst buffer must be a multiple of the block size.
	ReadBlocks(dst []byte, startBlockIdx int64) (int, error)
	// EraseBlocks erases blocks starting at startBlockIdx to startBlockIdx+numBlocks.
	EraseBlocks(startBlock, numBlocks int64) error
}

// NewBlockDevice creates a new BlockDevice from a Card.
func NewBlockDevice(card Card, blockSize int, numBlocks int64) (*BlockDevice, error) {
	if card == nil || blockSize <= 0 || numBlocks <= 0 {
		return nil, errors.New("invalid argument(s)")
	}
	blk, err := makeBlockIndexer(blockSize)
	if err != nil {
		return nil, err
	}
	bd := &BlockDevice{
		card:      card,
		blockbuf:  make([]byte, blockSize),
		blk:       blk,
		numblocks: int64(numBlocks),
	}
	return bd, nil
}

// BlockDevice implements tinyfs.BlockDevice interface for an [sd.Card] type.
type BlockDevice struct {
	card      Card
	blockbuf  []byte
	blk       blkIdxer
	numblocks int64
}

// ReadAt implements [io.ReadAt] interface for an SD card.
func (bd *BlockDevice) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errNegativeOffset
	}

	blockIdx := bd.blk.idx(off)
	blockOff := bd.blk.off(off)
	if blockOff != 0 {
		// Non-aligned first block case.
		if _, err = bd.card.ReadBlocks(bd.blockbuf, blockIdx); err != nil {
			return n, err
		}
		n += copy(p, bd.blockbuf[blockOff:])
		p = p[n:]
		blockIdx++
	}

	fullBlocksToRead := bd.blk.idx(int64(len(p)))
	if fullBlocksToRead > 0 {
		// 1 or more full blocks case.
		endOffset := fullBlocksToRead * bd.blk.size()
		ngot, err := bd.card.ReadBlocks(p[:endOffset], blockIdx)
		if err != nil {
			return n + ngot, err
		}
		p = p[endOffset:]
		n += ngot
		blockIdx += fullBlocksToRead
	}

	if len(p) > 0 {
		// Non-aligned last block case.
		if _, err := bd.card.ReadBlocks(bd.blockbuf, blockIdx); err != nil {
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

	blockIdx := bd.blk.idx(off)
	blockOff := bd.blk.off(off)
	if blockOff != 0 {
		// Non-aligned first block case.
		if _, err := bd.card.ReadBlocks(bd.blockbuf, blockIdx); err != nil {
			return n, err
		}
		nexpect := copy(bd.blockbuf[blockOff:], p)
		ngot, err := bd.card.WriteBlocks(bd.blockbuf, blockIdx)
		if err != nil {
			return n, err
		} else if ngot != len(bd.blockbuf) {
			return n, io.ErrShortWrite
		}
		n += nexpect
		p = p[nexpect:]
		blockIdx++
	}

	fullBlocksToWrite := bd.blk.idx(int64(len(p)))
	if fullBlocksToWrite > 0 {
		// 1 or more full blocks case.
		endOffset := fullBlocksToWrite * bd.blk.size()
		ngot, err := bd.card.WriteBlocks(p[:endOffset], blockIdx)
		n += ngot
		if err != nil {
			return n, err
		} else if ngot != int(endOffset) {
			return n, io.ErrShortWrite
		}
		p = p[ngot:]
		blockIdx += fullBlocksToWrite
	}

	if len(p) > 0 {
		// Non-aligned last block case.
		if _, err := bd.card.ReadBlocks(bd.blockbuf, blockIdx); err != nil {
			return n, err
		}
		copy(bd.blockbuf, p)
		ngot, err := bd.card.WriteBlocks(bd.blockbuf, blockIdx)
		if err != nil {
			return n, err
		} else if ngot != len(bd.blockbuf) {
			return n, io.ErrShortWrite
		}
		n += len(p)
	}
	return n, nil
}

// Size returns the number of bytes in this block device.
func (bd *BlockDevice) Size() int64 {
	return bd.BlockSize() * bd.numblocks
}

// BlockSize returns the size of a block in bytes.
func (bd *BlockDevice) BlockSize() int64 {
	return bd.blk.size()
}

// EraseBlocks erases the given number of blocks. An implementation may
// transparently coalesce ranges of blocks into larger bundles if the chip
// supports this. The start and len parameters are in block numbers, use
// EraseBlockSize to map addresses to blocks.
func (bd *BlockDevice) EraseBlocks(startEraseBlockIdx, len int64) error {
	return bd.card.EraseBlocks(startEraseBlockIdx, len)
}

// blkIdxer is a helper for calculating block indices and offsets.
type blkIdxer struct {
	blockshift int64
	blockmask  int64
}

func makeBlockIndexer(blockSize int) (blkIdxer, error) {
	if blockSize <= 0 {
		return blkIdxer{}, errNoblocks
	}
	tz := bits.TrailingZeros(uint(blockSize))
	if blockSize>>tz != 1 {
		return blkIdxer{}, errors.New("blockSize must be a power of 2")
	}
	blk := blkIdxer{
		blockshift: int64(tz),
		blockmask:  (1 << tz) - 1,
	}
	return blk, nil
}

// size returns the size of a block in bytes.
func (blk *blkIdxer) size() int64 {
	return 1 << blk.blockshift
}

// off gets the offset of the byte at byteIdx from the start of its block.
//
//go:inline
func (blk *blkIdxer) off(byteIdx int64) int64 {
	return blk._moduloBlockSize(byteIdx)
}

// idx gets the block index that contains the byte at byteIdx.
//
//go:inline
func (blk *blkIdxer) idx(byteIdx int64) int64 {
	return blk._divideBlockSize(byteIdx)
}

// modulo and divide are defined in terms of bit operations for speed since
// blockSize is a power of 2.

//go:inline
func (blk *blkIdxer) _moduloBlockSize(n int64) int64 { return n & blk.blockmask }

//go:inline
func (blk *blkIdxer) _divideBlockSize(n int64) int64 { return n >> blk.blockshift }
