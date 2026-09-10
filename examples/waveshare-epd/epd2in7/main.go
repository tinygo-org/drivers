// Slideshow demo for the Waveshare 2.7in e-Paper HAT V1 on a Waveshare
// RP2040-PiZero.
//
// Panel version.
//
// This demo is written for the V1 panel, which uses an IL91874-style
// controller. It uses the epd2in7 driver, which sends the matching init
// sequence and 5 part LUT (VCOM, WW, BW, BB, WB). Waveshare has not sold a
// V2 or later revision of the 2.7in panel, unlike the 2.13in and 2.9in
// panels, so this driver should match any 2.7in HAT panel.
//
//	V1 (IL91874-style, 176x264)  works, this demo
//
// Board support.
//
// The pin numbers below are only correct for the RP2040-PiZero. Any other
// board needs its own mapping.
//
//	Waveshare RP2040-PiZero          tested, works
//	Raspberry Pi Pico and Pico W     untested, needs its own pin mapping
//	                                 because it has no 40 pin header
//	Other RP2040 boards              untested, needs its own pin mapping
//
// The RP2040-PiZero has a Raspberry Pi 40 pin header. Waveshare exchanges
// GPIO10 and GPIO11 on the header so that RP2040 SPI1 SCK and TX align with
// the Raspberry Pi SCLK and MOSI positions.
// Schematic: https://files.waveshare.com/wiki/RP2040-PiZero/RP2040-PiZero.pdf
//
// Header to RP2040 GPIO for the e-Paper HAT signals:
//
//	pin 11 RST   GPIO17
//	pin 18 BUSY  GPIO24
//	pin 19 MOSI  GPIO11 (SPI1 SDO)
//	pin 22 DC    GPIO25
//	pin 23 SCLK  GPIO10 (SPI1 SCK)
//	pin 24 CS    GPIO8
//
// Images.
//
// Each file in assets/ is a packed 1 bit per pixel image, generated from a
// source picture by gen.py. The firmware embeds every asset with go:embed
// and cycles through them in name order, one every slideInterval. File
// names are numbered so the slideshow order is explicit.
//
//	assets/00-hello-mac.bin   the original 1984 Macintosh "hello" screen
//	assets/01-kanagawa.bin    The Great Wave off Kanagawa, Katsushika Hokusai
//	assets/02-cccp-stamp.bin  a 1972 Soviet stamp, 15 years of the space era
//	assets/03-guinea-pig.bin  a guinea pig photo
//
// The image is drawn at Rotation270 (Rotation90 landscape, plus 180 degrees
// for how this panel sits on its connector). Other wiring may need a
// different rotation to read right way up.
//
// Build and flash:
//
//	tinygo flash -target=pico ./examples/waveshare-epd/epd2in7
package main

import (
	"embed"
	"encoding/binary"
	"errors"
	"image/color"
	"machine"
	"sort"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/waveshare-epd/epd2in7"
)

//go:embed assets/*.bin
var assets embed.FS

const slideInterval = 10 * time.Second

var (
	black = color.RGBA{0, 0, 0, 255}
	white = color.RGBA{255, 255, 255, 255}

	display epd2in7.Device
)

func main() {
	// Wait for the USB serial console to attach so the log is not lost.
	time.Sleep(3 * time.Second)

	err := machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 4000000,
		SCK:       machine.GPIO10,
		SDO:       machine.GPIO11,
		SDI:       machine.GPIO12,
		Mode:      0,
	})
	if err != nil {
		println("SPI configure failed:", err.Error())
		return
	}

	display = epd2in7.New(machine.SPI1, machine.GPIO8, machine.GPIO25, machine.GPIO17, machine.GPIO24)
	display.Configure(epd2in7.Config{
		Rotation: drivers.Rotation270,
	})

	names, err := assetNames()
	if err != nil {
		println("reading assets failed:", err.Error())
		return
	}
	if len(names) == 0 {
		println("no assets embedded")
		return
	}

	println("epd2in7 slideshow: init done, clearing display")
	display.ClearBuffer()
	display.ClearDisplay()

	for i := 0; ; i = (i + 1) % len(names) {
		name := names[i]
		println("epd2in7 slideshow: showing", name)
		if err := showAsset(name); err != nil {
			println("showing", name, "failed:", err.Error())
			continue
		}
		time.Sleep(slideInterval)
	}
}

// assetNames returns the embedded asset file names in sorted order, so the
// slideshow order is deterministic and reproducible.
func assetNames() ([]string, error) {
	entries, err := assets.ReadDir("assets")
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, "assets/"+e.Name())
	}
	sort.Strings(names)
	return names, nil
}

// showAsset decodes one packed 1bpp asset file and draws it to the panel.
// The file format is a 4 byte header (uint16 width, uint16 height, little
// endian) followed by packed pixel bits, MSB first, row-major, bit=1 white.
func showAsset(name string) error {
	data, err := assets.ReadFile(name)
	if err != nil {
		return err
	}
	if len(data) < 4 {
		return errors.New("asset too short")
	}
	w := int16(binary.LittleEndian.Uint16(data[0:2]))
	h := int16(binary.LittleEndian.Uint16(data[2:4]))
	bits := data[4:]

	want := (int(w)*int(h) + 7) / 8
	if len(bits) < want {
		return errors.New("asset truncated")
	}

	for y := int16(0); y < h; y++ {
		for x := int16(0); x < w; x++ {
			bitpos := int(y)*int(w) + int(x)
			bit := bits[bitpos/8] & (0x80 >> uint(bitpos%8))
			if bit != 0 {
				display.SetPixel(x, y, white)
			} else {
				display.SetPixel(x, y, black)
			}
		}
	}

	return display.Display()
}
