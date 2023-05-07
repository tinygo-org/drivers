package neopixel

import (
	"fmt"
	"strconv"
	"time"
	"tinygo.org/x/drivers/seesaw"
)

// seesawWriteDelay the seesaw is quite timing sensitive and times out if not given enough time,
// this is an empirically determined delay that seems to have good results
const seesawWriteDelay = time.Millisecond * 10

// RGBW represents the RGBW color of an LED.
type RGBW struct {
	R, G, B, W uint8
}

type Device struct {
	seesaw          *seesaw.Device
	ledCount        int
	pin             uint8
	pixelType       PixelType
	lastOperationAt time.Time
}

func New(dev *seesaw.Device, pin uint8, ledCount int, pixelType PixelType) (*Device, error) {

	pixel := &Device{
		seesaw:    dev,
		ledCount:  ledCount,
		pin:       pin,
		pixelType: pixelType,
	}

	if !pixel.checkBufferLength(ledCount) {
		return nil, fmt.Errorf("invalid LED count: %d", ledCount)
	}

	time.Sleep(seesawWriteDelay)

	err := pixel.setupPin()
	if err != nil {
		return nil, fmt.Errorf("failed to update NeoPixel pin %d: %w", pin, err)
	}

	time.Sleep(seesawWriteDelay)

	err = pixel.setupLedCount()
	if err != nil {
		return nil, fmt.Errorf("failed to update LED count %d: %w", ledCount, err)
	}

	time.Sleep(seesawWriteDelay)

	err = pixel.setupSpeed()
	if err != nil {
		return nil, fmt.Errorf("failed to update pixel type: %w", err)
	}

	time.Sleep(seesawWriteDelay)

	return pixel, nil
}

func (s *Device) setupLedCount() error {

	lenBytes := calculateBufferLength(s.ledCount, s.pixelType)
	buf := []byte{byte(lenBytes >> 8), byte(lenBytes & 0xFF)}
	return s.seesaw.Write(seesaw.ModuleNeoPixelBase, seesaw.FunctionNeopixelBufLength, buf)
}

func calculateBufferLength(ledCount int, pixelType PixelType) int {
	return ledCount * pixelType.EncodedLen()
}

func (s *Device) setupSpeed() error {
	speed := byte(0)
	if s.pixelType.Is800KHz() {
		speed = 1
	}
	return s.seesaw.WriteRegister(seesaw.ModuleNeoPixelBase, seesaw.FunctionNeopixelSpeed, speed)
}

func (s *Device) setupPin() error {
	return s.seesaw.WriteRegister(seesaw.ModuleNeoPixelBase, seesaw.FunctionNeopixelPin, s.pin)
}

// WriteColorAtOffset updates the color for a single LED at the given offset
func (s *Device) WriteColorAtOffset(offset uint16, color RGBW) error {

	encodedLen := s.pixelType.EncodedLen()

	buf := make([]byte, encodedLen)
	l := s.pixelType.PutRGBW(buf, color)
	if l != encodedLen {
		panic("unexpected encoded length: " + strconv.Itoa(l) + " != " + strconv.Itoa(encodedLen))
	}
	byteOffset := offset * uint16(encodedLen)
	return s.writeBuffer(byteOffset, buf)
}

// WriteColors writes the given colors to the seesaws NeoPixel buffer
func (s *Device) WriteColors(buf []RGBW) error {

	if len(buf) > s.ledCount {
		return fmt.Errorf("buffer too big, only %d LEDs setup: %d > %d", s.ledCount, len(buf), s.ledCount)
	}

	encodedLen := s.pixelType.EncodedLen()

	tx := make([]byte, encodedLen*len(buf))
	pos := 0
	for _, c := range buf {
		w := tx[pos:]
		n := s.pixelType.PutRGBW(w, c)
		pos += n
	}

	// the seesaw can at most deal with 30 bytes according to the datasheet, but
	// crashes after 29 bytes. So we only send 29 data bytes at a time
	const chunkSize = 29

	// write the data chunk-by-chunk
	for i := 0; i < len(tx); i += chunkSize {
		toSend := tx[i:min(i+chunkSize, len(tx))]
		err := s.writeBuffer(uint16(i), toSend)
		if err != nil {
			return fmt.Errorf("failed to write NeoPixel buffer offset %d: %w", i, err)
		}
	}

	return nil
}

func (s *Device) writeBuffer(byteOffset uint16, buf []byte) error {
	tx := make([]byte, 2+len(buf))
	tx[0] = uint8(byteOffset >> 8)
	tx[1] = uint8(byteOffset)
	copy(tx[2:], buf)
	return s.seesaw.Write(seesaw.ModuleNeoPixelBase, seesaw.FunctionNeopixelBuf, tx)
}

func (s *Device) ShowPixels() error {

	// at most every 300us
	// https://github.com/adafruit/Adafruit_Seesaw/blob/8a2dc5e0645239cb34e23a4b62c456436b098ab3/seesaw_neopixel.cpp#L109
	s.waitSinceLastOperation(time.Microsecond * 300)

	return s.seesaw.Write(seesaw.ModuleNeoPixelBase, seesaw.FunctionNeopixelShow, nil)
}

func (s *Device) waitSinceLastOperation(d time.Duration) {
	diff := time.Since(s.lastOperationAt)
	for diff < d {
		time.Sleep(50 * time.Microsecond)
		diff = time.Since(s.lastOperationAt)
	}
}

// checkBufferLength checks whether the length is supported by seesaw. This depends on the pixel type.
// The seesaw has built in NeoPixel support for up to 170 RGB or 127 RGBW pixels. The
// output pin as well as the communication protocol frequency are configurable. Note:
// older firmware is limited to 63 pixels max.
func (s *Device) checkBufferLength(l int) bool {
	const maxRgbwPixelCount = 127
	const maxRgbPixelCount = 170

	if s.pixelType.IsRGBW() {
		return l <= maxRgbwPixelCount
	}

	return l <= maxRgbPixelCount
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
