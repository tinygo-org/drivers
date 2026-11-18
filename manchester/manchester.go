package manchester

import (
	"errors"
	"time"
)

// DefaultBitRate is used when Config.BitRate is zero. The SYN115 supports up
// to 10kbps and the SYN480R supports about 5kbps in sweep mode.
const DefaultBitRate = 2000

// DefaultPreamble is the number of 0x00 bytes sent before each frame when
// Config.Preamble is zero. The first bytes give a receiver time to settle.
const DefaultPreamble = 4

// preambleCheck is the number of preamble bytes the decoder validates.
const preambleCheck = 2

// Transmitter is the physical layer of a transmitter. Set turns the carrier
// on when high and off when low. The SYN115 data pin implements it.
type Transmitter interface {
	Set(high bool)
}

// Receiver is the physical layer of a receiver. Get returns the demodulated
// level, high when a carrier is present. The SYN480R data pin implements it.
type Receiver interface {
	Get() bool
}

// Config holds the settings shared by the encoder and the decoder.
type Config struct {
	// BitRate is the Manchester bit rate in bits per second.
	// Zero uses DefaultBitRate.
	BitRate uint32

	// Preamble is the number of 0x00 bytes sent before each frame.
	// Zero uses DefaultPreamble. The decoder checks the last two bytes.
	Preamble int

	// Timeout limits how long Read waits for a frame. Zero waits forever.
	Timeout time.Duration
}

// Errors returned by Decode and Decoder.Read.
var (
	ErrFrameLength = errors.New("manchester: waveform is not a whole number of bytes")
	ErrSymbol      = errors.New("manchester: invalid Manchester symbol")
	ErrTimeout     = errors.New("manchester: timeout")
)

func (c Config) bitPeriod() time.Duration {
	rate := c.BitRate
	if rate == 0 {
		rate = DefaultBitRate
	}
	return time.Second / time.Duration(rate)
}

func (c Config) preambleBytes() int {
	if c.Preamble <= 0 {
		return DefaultPreamble
	}
	return c.Preamble
}

// Encode converts data to Manchester half-bit levels. Bit 0 is a high level
// followed by a low level, bit 1 is a low level followed by a high level.
// Bits are sent most significant bit first.
func Encode(data []byte) []bool {
	wave := make([]bool, 0, len(data)*16)
	for _, b := range data {
		for bit := 7; bit >= 0; bit-- {
			if b&(1<<bit) != 0 {
				wave = append(wave, false, true)
			} else {
				wave = append(wave, true, false)
			}
		}
	}
	return wave
}

// Decode converts Manchester half-bit levels to bytes. The length must be a
// whole number of bytes, so each byte takes 16 levels.
func Decode(wave []bool) ([]byte, error) {
	if len(wave)%16 != 0 {
		return nil, ErrFrameLength
	}
	data := make([]byte, len(wave)/16)
	for i := range data {
		var b byte
		for bit := 7; bit >= 0; bit-- {
			first := wave[16*i+2*(7-bit)]
			second := wave[16*i+2*(7-bit)+1]
			switch {
			case first && !second:
			case !first && second:
				b |= 1 << bit
			default:
				return nil, ErrSymbol
			}
		}
		data[i] = b
	}
	return data, nil
}

// Encoder sends Manchester frames through a Transmitter.
type Encoder struct {
	tx       Transmitter
	half     time.Duration
	preamble int
}

// NewEncoder returns an Encoder that drives tx.
func NewEncoder(tx Transmitter, config Config) *Encoder {
	return &Encoder{
		tx:       tx,
		half:     config.bitPeriod() / 2,
		preamble: config.preambleBytes(),
	}
}

// Write sends p as one Manchester frame. It returns len(p) or an error.
func (e *Encoder) Write(p []byte) (int, error) {
	deadline := time.Now()
	for i := 0; i < e.preamble; i++ {
		e.writeByte(&deadline, 0)
	}
	for _, b := range p {
		e.writeByte(&deadline, b)
	}
	// Idle low so the next frame starts with a rising edge.
	e.tx.Set(false)
	return len(p), nil
}

func (e *Encoder) writeByte(deadline *time.Time, b byte) {
	for bit := 7; bit >= 0; bit-- {
		if b&(1<<bit) != 0 {
			e.writeLevel(deadline, false)
			e.writeLevel(deadline, true)
		} else {
			e.writeLevel(deadline, true)
			e.writeLevel(deadline, false)
		}
	}
}

func (e *Encoder) writeLevel(deadline *time.Time, high bool) {
	*deadline = deadline.Add(e.half)
	e.tx.Set(high)
	if d := time.Until(*deadline); d > 0 {
		time.Sleep(d)
	}
}

// Decoder receives Manchester frames from a Receiver.
type Decoder struct {
	rx       Receiver
	half     time.Duration
	poll     time.Duration
	preamble int
	timeout  time.Duration
}

// NewDecoder returns a Decoder that reads rx.
func NewDecoder(rx Receiver, config Config) *Decoder {
	half := config.bitPeriod() / 2
	poll := half / 8
	if poll < time.Microsecond {
		poll = time.Microsecond
	}
	return &Decoder{
		rx:       rx,
		half:     half,
		poll:     poll,
		preamble: config.preambleBytes(),
		timeout:  config.Timeout,
	}
}

// Read waits for one frame and decodes it into p. It returns len(p) on
// success. Partial frames are discarded and reading continues until the
// timeout expires.
func (d *Decoder) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	var deadline time.Time
	if d.timeout > 0 {
		deadline = time.Now().Add(d.timeout)
	}

	for {
		if err := d.waitFor(false, deadline); err != nil {
			return 0, err
		}
		if err := d.waitFor(true, deadline); err != nil {
			return 0, err
		}

		wave := d.read(2*8*d.preamble + 16*len(p))
		if d.hasPreamble(wave) {
			data, err := Decode(wave[2*8*d.preamble:])
			if err == nil {
				copy(p, data)
				return len(p), nil
			}
		}

		if !deadline.IsZero() && time.Now().After(deadline) {
			return 0, ErrTimeout
		}
	}
}

func (d *Decoder) waitFor(high bool, deadline time.Time) error {
	for {
		if d.rx.Get() == high {
			return nil
		}
		if !deadline.IsZero() && time.Now().After(deadline) {
			return ErrTimeout
		}
		time.Sleep(d.poll)
	}
}

func (d *Decoder) read(n int) []bool {
	wave := make([]bool, n)
	if n == 0 {
		return wave
	}
	// The rising edge that started the frame is the first half-bit.
	wave[0] = true
	next := time.Now().Add(d.half + d.half/2)
	for i := 1; i < len(wave); i++ {
		if wait := time.Until(next); wait > 0 {
			time.Sleep(wait)
		}
		wave[i] = d.rx.Get()
		next = next.Add(d.half)
	}
	return wave
}

func (d *Decoder) hasPreamble(wave []bool) bool {
	check := d.preamble
	if check > preambleCheck {
		check = preambleCheck
	}
	start := 2 * 8 * (d.preamble - check)
	for i := start; i < 2*8*d.preamble; i += 2 {
		if !wave[i] || wave[i+1] {
			return false
		}
	}
	return true
}
