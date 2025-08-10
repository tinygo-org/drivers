package sharpmem

import (
	"image/color"
	"math/rand/v2"
	"testing"

	qt "github.com/frankban/quicktest"
)

func Test_setBit(t *testing.T) {
	c := qt.New(t)

	for i := uint8(0); i < 8; i++ {
		v := uint8(1) << i

		c.Assert(setBit(0x00, i), qt.Equals, v)
		c.Assert(setBit(0x00, (i+1)%8), qt.Not(qt.Equals), v)
	}
}

func Test_unsetBit(t *testing.T) {
	c := qt.New(t)

	for i := uint8(0); i < 8; i++ {
		v := uint8(1) << i

		c.Assert(unsetBit(v, i), qt.Equals, uint8(0x00))
		c.Assert(unsetBit(v, (i+1)%8), qt.Not(qt.Equals), uint8(0x00))
	}
}

func Test_hasBit(t *testing.T) {
	c := qt.New(t)

	for i := uint8(0); i < 8; i++ {
		v := uint8(1) << i

		c.Assert(hasBit(v, i), qt.Equals, true)
		c.Assert(hasBit(v, (i+1)%8), qt.Equals, false)
	}
}

func Test_bitfieldBufLen(t *testing.T) {
	c := qt.New(t)

	for i := int16(1); i < 536; i++ {
		requiredBufferSize := i / 8
		wouldOverflow := i % 8

		if wouldOverflow > 0 {
			requiredBufferSize += 1
		}

		c.Assert(bitfieldBufLen(i), qt.Equals, requiredBufferSize)
	}
}

type mockBus struct {
	b []byte
}

func (m *mockBus) Tx(w, _ []byte) error {
	m.b = append(m.b, w...)
	return nil
}

func (m *mockBus) Transfer(b byte) (byte, error) {
	m.b = append(m.b, b)
	return 0x00, nil
}

type mockPin struct{}

func (m mockPin) High() {
}

func (m mockPin) Low() {
}

func Test_Device(t *testing.T) {
	c := qt.New(t)

	cfgs := []Config{
		ConfigLS010B7DH04,
		ConfigLS011B7DH03,
		ConfigLS012B7DD01,
		ConfigLS013B7DH03,
		ConfigLS013B7DH05,
		ConfigLS018B7DH02,
		ConfigLS027B7DH01,
		ConfigLS027B7DH01A,
		ConfigLS032B7DD02,
		ConfigLS044Q7DH01,
	}

	cfgLen := len(cfgs)
	for i := 0; i < cfgLen; i++ {
		cfgs = append(cfgs, Config{
			Width:                cfgs[i].Width,
			Height:               cfgs[i].Height,
			DisableOptimizations: true,
		})
	}

	spi := &mockBus{}
	pin := mockPin{}
	display := New(spi, pin)

	for _, cfg := range cfgs {
		display.Configure(cfg)

		x, y := display.Size()
		c.Assert(x, qt.Equals, cfg.Width)
		c.Assert(y, qt.Equals, cfg.Height)

		for i := 0; i < 10; i++ {
			x := int16(rand.IntN(int(cfg.Width)))
			y := int16(rand.IntN(int(cfg.Height)))
			display.SetPixel(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}

		for i := 0; i < 10; i++ {
			x := int16(rand.IntN(int(cfg.Width)))
			y := int16(rand.IntN(int(cfg.Height)))
			display.SetPixel(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
		}

		err := display.Display()
		c.Assert(err, qt.Equals, nil)

		err = display.ClearDisplay()
		c.Assert(err, qt.Equals, nil)

		display.ClearBuffer()
	}
}

func Test_HiPad(t *testing.T) {
	c := qt.New(t)

	spi := &mockBus{}
	pin := mockPin{}
	display := New(spi, pin)

	t.Run("LS011B7DH03, 8-bit address", func(t *testing.T) {
		t.Cleanup(func() {
			spi.b = nil
		})

		display.Configure(ConfigLS011B7DH03)

		display.SetPixel(0, display.height-1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		err := display.Display()
		c.Assert(err, qt.Equals, nil)

		// 160 perfectly divisible by 16, so 20 bytes of pixel data
		c.Assert(spi.b, qt.HasLen, 2+20+2)

		// line is 1-indexed on the wire (67+1)
		// 68 in binary
		// 0b01000100

		//                                    DDDDDMMM
		c.Assert(spi.b[0], qt.Equals, uint8(0b00000011)) // mode 1, vcom is high on first run

		c.Assert(spi.b[1], qt.Equals, uint8(0b01000100)) // the actual address
	})

	t.Run("LS018B7DH02, 9-bit address", func(t *testing.T) {
		t.Cleanup(func() {
			spi.b = nil
		})

		display.Configure(ConfigLS018B7DH02)

		display.SetPixel(0, display.height-1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		err := display.Display()
		c.Assert(err, qt.Equals, nil)

		// 2 first bytes command+address
		// 230 bits are not divisible by 16, 240 is (15*16), so 30 bytes for line data
		// 2 trailing bytes
		c.Assert(spi.b, qt.HasLen, 2+30+2)

		// line is 1-indexed on the wire (302+1)
		// 303 in binary (split in 2 bytes)
		//          R
		// 0b00000001 0b00101111
		//          ^

		//                                    RDDDDMMM
		c.Assert(spi.b[0], qt.Equals, uint8(0b10000011)) // mode 1, vcom is high on first run
		//                                    ^

		c.Assert(spi.b[1], qt.Equals, uint8(0b00101111)) // rest of the address (low 8 bits)
	})

	t.Run("LS032B7DD02, 10-bit address", func(t *testing.T) {
		t.Cleanup(func() {
			spi.b = nil
		})

		display.Configure(ConfigLS032B7DD02)

		display.SetPixel(0, display.height-1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		err := display.Display()
		c.Assert(err, qt.Equals, nil)

		c.Assert(spi.b, qt.HasLen, 2+336/8+2) // 2 command+address, width / 2, 2 trailing bytes

		// line is 1-indexed on the wire (535+1)
		// 536 in binary (split in 2 bytes)
		//         RR
		// 0b00000010 0b00011000
		//         ^^

		//                                    RRDDDMMM
		c.Assert(spi.b[0], qt.Equals, uint8(0b10000011)) // mode 1, vcom is high on first run
		//                                    ^^

		c.Assert(spi.b[1], qt.Equals, uint8(0b00011000)) // rest of the address (low 8 bits)
	})

}
