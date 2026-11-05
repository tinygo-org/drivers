# Waveshare 2.7in four-shade e-paper

This driver supports the 176x264 V1 panel's four shades, black, dark gray,
light gray, and white. Waveshare calls this mode `4Gray`. It uses **2 bits per
pixel**, not 4 bits or 16 shades. The monochrome `epd2in7` driver is unchanged.

Configure the SPI bus and GPIO pins before calling `New`. CS, DC, and RST
must be outputs and BUSY must be an input. The pin interfaces accept
`machine.Pin` and allow host-side testing without the `machine` package.

```go
display := epd2in7gray.New(spi, cs, dc, rst, busy)
if err := display.Configure(epd2in7gray.Config{
    Rotation: drivers.Rotation270,
}); err != nil {
    return err
}
display.SetPixel(10, 10, color.RGBA{85, 85, 85, 255})
return display.Display()
```

`SetPixel` converts RGB luminance to four shades and ignores alpha.
The framebuffer uses 11,616 bytes at native resolution. Refreshing streams
two 5,808-byte bitplanes using a 64-byte scratch buffer. `LogicalWidth` can
add row padding without changing the visible `Size` or transmitting padding.
SPI uses one byte per CS assertion, matching the vendor and monochrome drivers.

`Display` blocks for a full refresh. Partial refresh and dithering are not
implemented. `ClearBuffer` changes buffered pixels to white without I/O.
`ClearDisplay` refreshes the panel to white without changing those pixels.
SPI errors and BUSY timeouts are returned to the caller. `BusyTimeout`
defaults to 30 seconds and limits each BUSY wait, not the entire operation.

`DeepSleep` retains the displayed image. Call `Configure` to wake the panel
and restore its grayscale settings. Configuration clears the framebuffer.

Both grayscale examples use the HAT's 8-pin connector with a Waveshare
RP2040-PiZero. The pin mapping matches the monochrome `epd2in7` example.
Disconnect power before connecting the wires.

| HAT label | RP2040-PiZero header pin | RP2040 GPIO |
| --- | --- | --- |
| VCC | 3.3V | — |
| GND | GND | — |
| DIN | 19 | GP11, SPI1 TX |
| CLK | 23 | GP10, SPI1 SCK |
| CS | 24 | GP8 |
| DC | 22 | GP25 |
| RST | 11 | GP17 |
| BUSY | 18 | GP24 |

SPI1 uses 4 MHz and mode 0. MISO is not connected. Use 3.3 V power and logic.
See the [Waveshare connector table](https://www.waveshare.com/wiki/2.7inch_e-Paper_HAT_Manual#Hardware_Connection)
and [RP2040-PiZero pinout](https://www.waveshare.com/wiki/RP2040-PiZero).

```sh
tinygo flash -target=waveshare-rp2040-pizero ./examples/waveshare-epd/epd2in7gray
```

The example draws four labeled swatches and a top-left orientation marker.
It prints each initialization stage without diagnostic pauses.
The driver's optional `OnInitStage` callback enables initialization tracing.
Without a callback, the driver adds no diagnostic pauses or console output.
Check the shade order and distinction between the grays on a real panel.
Software tests cannot verify physical waveform quality.

The separate [slideshow example](../../examples/waveshare-epd/epd2in7gray-slideshow)
uses packed 2bpp images with the original slideshow's width/height header.
Its [asset format](../../examples/waveshare-epd/epd2in7gray-slideshow/assets/README.md)
describes how to add images. The slideshow has no diagnostic pauses.

The controller sequence and waveform tables come from
[Waveshare's driver](https://github.com/waveshareteam/e-Paper/blob/a824b4f8f34dee7e721183c0154788dcde41c460/Arduino_R4/src/e-Paper/EPD_2in7.cpp).
The applicable vendor license is included in this directory.
