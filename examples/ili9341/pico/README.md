# Adafruit 2.2" TFT with Raspberry Pi Pico and Pico 2

This example shows red, green, blue, and white bars on the
[Adafruit 2.2" 240x320 TFT display, product 1480](https://www.adafruit.com/product/1480).
It uses the `ili9341` driver with the display's ILI9340C controller.

The example was tested on a Pico 2 W. It also builds for Pico, Pico W, and
Pico 2, but these boards have not been tested with this display.

## Wiring

Disconnect USB power before you connect the wires. The physical pin numbers
are the same on Pico, Pico W, Pico 2, and Pico 2 W.

| Display pin | Pico pin | Physical pin |
|---|---|---|
| `Vin` | `VBUS` | 40 |
| `GND` | `GND` | 38 |
| `CLK` | `GP18` | 24 |
| `MOSI` | `GP19` | 25 |
| `LCD_CS` | `GP17` | 22 |
| `D/C` | `GP20` | 26 |
| `RESET` | `GP21` | 27 |

This wiring uses USB power. `VBUS` supplies 5 V to the display's `Vin` pin.
Do not connect 5 V to a Pico GPIO pin.

Leave `Backlite`, `MISO`, and `SDCS` unconnected. This example does not use the
microSD card slot.

See the [Adafruit pin descriptions](https://learn.adafruit.com/2-2-tft-display/pinouts).

## Flash

From the repository root, use the command for your board:

| Board | Command |
|---|---|
| Pico | `tinygo flash -target=pico ./examples/ili9341/pico` |
| Pico W | `tinygo flash -target=pico-w ./examples/ili9341/pico` |
| Pico 2 | `tinygo flash -target=pico2 ./examples/ili9341/pico` |
| Pico 2 W | `tinygo flash -target=pico2-w ./examples/ili9341/pico` |

Hold BOOTSEL while you connect USB, then release BOOTSEL before you run the
command. The display should show four horizontal bars, with red at the top
and white at the bottom.
