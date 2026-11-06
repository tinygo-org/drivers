# Arduino UNO R4 WiFi LED matrix

Driver for the built-in 12x8 monochrome matrix. Import
`tinygo.org/x/drivers/unor4matrix`. The driver implements `drivers.Displayer`.

## Usage

```go
display := unor4matrix.NewOnboard()
if err := display.Configure(); err != nil {
    return
}
display.SetPixel(0, 0, color.RGBA{R: 255})
for {
    if err := display.Display(); err != nil {
        return
    }
}
```

`NewOnboard` is available with `arduino_uno_r4_wifi`. `New` accepts eleven
`CharlieplexPin` values in this order:

```text
P0_03 P0_04 P0_11 P0_12 P0_13 P0_15 P2_04 P2_05 P2_06 P2_12 P2_13
```

Set must drive the requested level. Float must initialize a high-impedance input
without pull-ups. Configure returns `ErrMissingPinFunc` if either function is
missing, otherwise it floats every line. Configure may be repeated and keeps
the image. Give the device exclusive access to its pins.

## Refresh model

The image occupies 12 bytes. SetPixel edits it, ClearDisplay clears it, and
LoadFrame replaces it with Arduino-format words. Words are row-major, MSB first:
word 0 bit 31 is (0,0), word 2 bit 0 is (11,7). Any nonzero RGB is on, alpha is
ignored, and writes outside the drawing area are ignored.

Display scans all 96 slots directly from the image and floats each active pair
after its slot. Call it repeatedly to keep an image visible. Every slot sleeps
for 100 us, including dark pixels. A frame takes at least 9.6 ms plus GPIO and
scheduler overhead. Before successful Configure, Display returns ErrNotConfigured.

An active LED stays driven during time.Sleep. Other goroutines can prolong its on-time.

## Examples and limitations

```sh
tinygo flash -target=arduino-uno-r4-wifi ./examples/unor4matrix/checkerboard
tinygo flash -target=arduino-uno-r4-wifi ./examples/unor4matrix/life
```

Life runs a glider on a wrapping 12x8 grid with 300 ms generations. Display leaves
the lines floating before Life computes the next generation.

Use the default `tasks` scheduler. With TinyGo 0.44 and `-scheduler=none`, short
sleeps can wait for a 1 ms SysTick, reducing refresh to about 10 Hz. Display
blocks and foreground work pauses refresh. Call the driver from one application
goroutine, never from an interrupt handler. Rotation and brightness control are
unsupported. UNO R4 Minima has no built-in matrix. UNO Q uses `unoqmatrix`.

## Sources

- [Arduino mapping and 10 kHz timer](https://github.com/arduino/ArduinoCore-renesas/blob/424e86eff92d37f72123c2b641dd8bbf06a38b47/libraries/Arduino_LED_Matrix/src/Arduino_LED_Matrix.h).
- [UNO R4 WiFi line order](https://github.com/arduino/ArduinoCore-renesas/blob/424e86eff92d37f72123c2b641dd8bbf06a38b47/variants/UNOWIFIR4/variant.cpp).
- [RA4M1 hardware manual](https://www.renesas.com/en/document/mah/renesas-ra4m1-group-users-manual-hardware?language=en), R01UH0887EJ0110 Rev.1.10, sections 19.2.1, 19.2.3 and 19.2.5, pp. 357, 359 and 361-362. PODR holds the output level, POSR/PORR update it, and PDR enables output.
