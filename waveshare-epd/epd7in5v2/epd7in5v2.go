package epd7in5v2

import (
	"machine"
	"time"
)

// Waveshare 7.5 inch e-ink display v2
// https://www.waveshare.com/product/displays/e-paper/epaper-1/7.5inch-e-paper-hat.htm
//
// Currently only supporting 2 color display, and full display updates.
//
// https://files.waveshare.com/upload/6/60/7.5inch_e-Paper_V2_Specification.pdf
// https://github.com/waveshareteam/e-Paper/tree/4822c075f5df714f88b02e10c336b4eeff7e603e/Arduino/epd7in5_V2

const WIDTH int = 800
const HEIGHT int = 480
const BYTE_WIDTH int = WIDTH / 8

type EPD7in5_V2 struct {
	cs  machine.Pin
	rst machine.Pin
	dc  machine.Pin
	bsy machine.Pin
	pwr machine.Pin
	bus machine.SPI
}

func New(cs, rst, dc, bsy, pwr machine.Pin, bus machine.SPI) *EPD7in5_V2 {
	return &EPD7in5_V2{
		cs:  cs,
		rst: rst,
		dc:  dc,
		bsy: bsy,
		pwr: pwr,
		bus: bus,
	}
}

func (e *EPD7in5_V2) Configure(sck, sdo machine.Pin) error {
	e.bsy.Configure(machine.PinConfig{Mode: machine.PinInput})
	e.cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	e.rst.Configure(machine.PinConfig{Mode: machine.PinOutput})
	e.dc.Configure(machine.PinConfig{Mode: machine.PinOutput})
	e.pwr.Configure(machine.PinConfig{Mode: machine.PinOutput})
	e.pwr.High()

	return e.bus.Configure(machine.SPIConfig{
		Frequency: 2_000_000,
		SCK:       sck,
		SDO:       sdo,
		LSBFirst:  false,
		Mode:      0,
	})
}

func (e *EPD7in5_V2) Command(cmd byte, data []byte) error {
	e.dc.Low()
	e.cs.Low()
	_, err := e.bus.Transfer(cmd)
	e.cs.High()

	if err != nil || data == nil {
		return err
	}

	return e.Data(data)
}

func (e *EPD7in5_V2) Data(data []byte) error {
	e.dc.High()
	e.cs.Low()
	err := e.bus.Tx(data, []byte{})
	e.cs.High()
	return err
}

func (e *EPD7in5_V2) waitUntilIdle() error {
	var err error
	for {
		if err = e.Command(0x71, nil); err != nil {
			return err
		}
		if e.bsy.Get() {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(20 * time.Millisecond)
	return nil
}

func (e *EPD7in5_V2) Reset() {
	e.rst.High()
	time.Sleep(20 * time.Millisecond)
	e.rst.Low()
	time.Sleep(20 * time.Millisecond)
	e.rst.High()
	time.Sleep(20 * time.Millisecond)
}

func (e *EPD7in5_V2) Init() error {
	var err error

	// power setting
	if err = e.Command(0x01, []byte{0x17, 0x17, 0x3F, 0x3F, 0x11}); err != nil {
		return err
	}

	// VCOM DC setting
	if err = e.Command(0x82, []byte{0x24}); err != nil {
		return err
	}

	// booster setting
	if err = e.Command(0x06, []byte{0x27, 0x27, 0x2F, 0x17}); err != nil {
		return err
	}

	// OSC setting
	if err = e.Command(0x30, []byte{0x06}); err != nil {
		return err
	}

	// power on
	if err = e.Command(0x04, nil); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	if err = e.waitUntilIdle(); err != nil {
		return err
	}

	// panel setting
	if err = e.Command(0x00, []byte{0x3F}); err != nil {
		return err
	}

	// resolution setting
	if err = e.Command(0x61, []byte{0x03, 0x20, 0x01, 0xE0}); err != nil {
		return err
	}

	// dual SPI mode (off)
	if err = e.Command(0x15, []byte{0x00}); err != nil {
		return err
	}

	// VCOM and data interval
	if err = e.Command(0x50, []byte{0x10, 0x00}); err != nil {
		return err
	}

	// tcon setting
	if err = e.Command(0x60, []byte{0x22}); err != nil {
		return err
	}

	// Gate/Source Start Setting
	if err = e.Command(0x65, []byte{0x00, 0x00, 0x00, 0x00}); err != nil {
		return err
	}

	// Set LUT
	if err = e.Command(0x20, []byte{
		0x00, 0x0F, 0x0F, 0x00, 0x00, 0x01, 0x00, 0x0F, 0x01, 0x0F, 0x01, 0x02, 0x00, 0x0F,
		0x0F, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}); err != nil {
		return err
	}

	if err = e.Command(0x21, []byte{
		0x10, 0x0F, 0x0F, 0x00, 0x00, 0x01, 0x84, 0x0F, 0x01, 0x0F, 0x01, 0x02, 0x20, 0x0F,
		0x0F, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}); err != nil {
		return err
	}

	if err = e.Command(0x22, []byte{
		0x10, 0x0F, 0x0F, 0x00, 0x00, 0x01, 0x84, 0x0F, 0x01, 0x0F, 0x01, 0x02, 0x20, 0x0F,
		0x0F, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}); err != nil {
		return err
	}

	if err = e.Command(0x23, []byte{
		0x80, 0x0F, 0x0F, 0x00, 0x00, 0x01, 0x84, 0x0F, 0x01, 0x0F, 0x01, 0x02, 0x40, 0x0F,
		0x0F, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}); err != nil {
		return err
	}

	return e.Command(0x24, []byte{
		0x80, 0x0F, 0x0F, 0x00, 0x00, 0x01, 0x84, 0x0F, 0x01, 0x0F, 0x01, 0x02, 0x40, 0x0F,
		0x0F, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	})
}

func (e *EPD7in5_V2) Draw(img []byte) error {
	var err error
	if err = e.Command(0x13, img); err != nil {
		return err
	}
	if err = e.Command(0x12, nil); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return e.waitUntilIdle()
}

func (e *EPD7in5_V2) Clear() error {
	return e.Draw(make([]byte, BYTE_WIDTH*HEIGHT))
}

func (e *EPD7in5_V2) DeepSleep() error {
	return e.Command(0x07, nil)
}

func NewImageBuffer() []byte {
	return make([]byte, BYTE_WIDTH*HEIGHT)
}
