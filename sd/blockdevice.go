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
	WriteBlocks(data []byte, startBlockIdx int64) error
	ReadBlocks(dst []byte, startBlockIdx int64) error
	EraseBlocks(start, len int64) error
}

func NewBlockDevice(card Card, blockSize int, numBlocks, eraseBlockSize int64) (*BlockDevice, error) {
	if card == nil || blockSize <= 0 || eraseBlockSize <= 0 || numBlocks <= 0 {
		return nil, errors.New("invalid argument(s)")
	}
	tz := bits.TrailingZeros(uint(blockSize))
	if blockSize>>tz != 1 {
		return nil, errors.New("blockSize must be a power of 2")
	}
	bd := &BlockDevice{
		card:       card,
		blockbuf:   make([]byte, blockSize),
		blockshift: tz,
		blockmask:  (1 << tz) - 1,
		numblocks:  numBlocks,
	}
	return bd, nil
}

// BlockDevice implements tinyfs.BlockDevice interface.
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

func (bd *BlockDevice) Size() int64 {
	return int64(len(bd.blockbuf)) * bd.numblocks
}

func (bd *BlockDevice) EraseBlocks(start, len int64) error {
	return bd.card.EraseBlocks(start, len)
}

func (bd *BlockDevice) EraseBlockSize() int64 {
	return bd.eraseBlockSize
}
