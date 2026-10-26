//go:build tinygo

// Four-shade slideshow for the 2.7in V1 HAT on a Waveshare RP2040-PiZero.
// Flash with tinygo flash -target=waveshare-rp2040-pizero ./examples/waveshare-epd/epd2in7gray-slideshow
package main

import (
	"embed"
	"encoding/binary"
	"errors"
	"image/color"
	"io/fs"
	"machine"
	"path"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/waveshare-epd/epd2in7gray"
)

//go:embed assets/*
var assets embed.FS

const slideInterval = 10 * time.Second

func main() {
	time.Sleep(3 * time.Second)
	if err := run(); err != nil {
		println("epd2in7gray slideshow:", err.Error())
	}
}

func run() error {
	names, err := assetNames(assets)
	if err != nil {
		return err
	}
	if len(names) == 0 {
		return errors.New("no 2bpp images embedded; add .bin files to assets")
	}

	if err := machine.SPI1.Configure(machine.SPIConfig{
		Frequency: 4_000_000,
		SCK:       machine.GPIO10,
		SDO:       machine.GPIO11,
		SDI:       machine.GPIO12,
		Mode:      0,
	}); err != nil {
		return err
	}
	cs, dc, rst, busy := machine.GPIO8, machine.GPIO25, machine.GPIO17, machine.GPIO24
	for _, pin := range []machine.Pin{cs, dc, rst} {
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
	busy.Configure(machine.PinConfig{Mode: machine.PinInput})
	display := epd2in7gray.New(machine.SPI1, cs, dc, rst, busy)
	if err := display.Configure(epd2in7gray.Config{Rotation: drivers.Rotation270}); err != nil {
		return err
	}
	println("epd2in7gray slideshow: init done, clearing display")
	if err := display.ClearDisplay(); err != nil {
		return err
	}

	for i := 0; ; i = (i + 1) % len(names) {
		name := names[i]
		println("epd2in7gray slideshow: showing", name)
		data, err := assets.ReadFile(name)
		if err == nil {
			err = showAsset(display, data)
		}
		if err != nil {
			println("showing", name, "failed:", err.Error())
		}
		time.Sleep(slideInterval)
	}
}

var (
	errAssetHeader     = errors.New("asset header is missing")
	errAssetDimensions = errors.New("asset dimensions do not match display")
	errAssetLength     = errors.New("asset pixel data length does not match dimensions")
)

func assetNames(assets fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(assets, "assets")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && path.Ext(entry.Name()) == ".bin" {
			names = append(names, "assets/"+entry.Name())
		}
	}
	return names, nil
}

func showAsset(display drivers.Displayer, data []byte) error {
	if len(data) < 4 {
		return errAssetHeader
	}
	w := int(binary.LittleEndian.Uint16(data[:2]))
	h := int(binary.LittleEndian.Uint16(data[2:4]))
	displayWidth, displayHeight := display.Size()
	if w == 0 || h == 0 || w != int(displayWidth) || h != int(displayHeight) {
		return errAssetDimensions
	}
	pixels := data[4:]
	if len(pixels) != (w*h+3)/4 {
		return errAssetLength
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			index := y*w + x
			shift := uint(6 - (index%4)*2)
			gray := ((pixels[index/4] >> shift) & 3) * 85
			display.SetPixel(int16(x), int16(y), color.RGBA{gray, gray, gray, 255})
		}
	}
	return display.Display()
}
