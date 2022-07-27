package makeybutton

const bufferSize = 6

// Buffer is a buffer to keep track of the most recent readings for a button.
type Buffer struct {
	readings [bufferSize]bool
	index    int
}

// NewBuffer returns a new buffer.
func NewBuffer() *Buffer {
	return &Buffer{}
}

// Used returns how many bytes in buffer have been used.
func (b *Buffer) Used() int {
	return b.index
}

// Put stores a boolean in the buffer.
func (b *Buffer) Put(val bool) bool {
	b.index++
	if b.index >= bufferSize {
		b.index = 0
	}

	b.readings[b.index] = val

	return true
}

// Avg returns the "average" of all the readings in the buffer, by
// treating a true as 1 and a false as -1.
func (b *Buffer) Avg() int {
	avg := 0
	for i := 0; i < bufferSize; i++ {
		if b.readings[i] {
			avg += 1
		} else {
			avg -= 1
		}
	}

	return avg
}
