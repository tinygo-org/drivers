package makeybutton

const (
	bufferSize    = 3
	maxSumAllowed = 4
)

// Buffer is a buffer to keep track of the most recent readings for a button.
// in bit form.
type Buffer struct {
	data        [bufferSize]byte
	byteCounter int
	bitCounter  int
	sum         int
}

// NewBuffer returns a new buffer.
func NewBuffer() *Buffer {
	return &Buffer{}
}

// Sum returns the sum of all measurements
func (b *Buffer) Sum() int {
	return b.sum
}

// Put stores a boolean button state into the buffer.
func (b *Buffer) Put(val bool) {
	currentMeasurement, oldestMeasurement := b.updateData(val)
	b.updateCounters()

	if currentMeasurement != 0 && b.sum < maxSumAllowed {
		b.sum++
	}

	if oldestMeasurement != 0 && b.sum > 0 {
		b.sum--
	}
}

func (b *Buffer) updateData(val bool) (byte, byte) {
	currentByte := b.data[b.byteCounter]
	oldestMeasurement := (currentByte >> b.bitCounter) & 0x01

	if val {
		currentByte |= (1 << b.bitCounter)
	} else {
		currentByte &= ^(1 << b.bitCounter)
	}

	b.data[b.byteCounter] = currentByte

	return (currentByte >> b.bitCounter) & 0x01, oldestMeasurement
}

func (b *Buffer) updateCounters() {
	b.bitCounter++
	if b.bitCounter == 8 {
		b.bitCounter = 0
		b.byteCounter++
		if b.byteCounter == bufferSize {
			b.byteCounter = 0
		}
	}
}
