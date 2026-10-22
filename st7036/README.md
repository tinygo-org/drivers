# ST7036 character LCD

This package supports the Sitronix ST7036 controller on the Pimoroni
Display-o-Tron HAT and Display-o-Tron 3000. These boards have 16 columns
and 3 rows.

Configure the SPI bus before you call `New`. Use SPI mode 0 with a clock
frequency of 1 MHz. Supply chip-select, register-select, and reset pins.
Use `machine.NoPin` if reset is not connected. The driver configures its
output pins and sends each byte with a separate chip-select pulse.

Call `Configure(Config{})` to select 16 columns and 3 rows, clear the
display, and set contrast to 40. `Config` also accepts 1-row and 2-row
layouts. Call `SetCursorPosition(column, row)` to select a position.
Coordinates start at zero.

`Write` implements `io.Writer` and sends bytes directly to the display.
It does not buffer text or process newlines. Use `CreateChar` to define
slots 0 to 7, then write a byte with the slot number. `SetDisplayMode`
controls the display, cursor, and cursor blink.

The command sequence and delays follow the
[Pimoroni driver](https://github.com/pimoroni/st7036).
The driver does not return SPI transfer errors. The package tests do not
verify hardware behavior.

Run package tests with `go test -v ./st7036` or `tinygo test -v ./st7036`.
These tests use mock SPI and GPIO and do not need hardware. The `New`
constructor uses `machine.Pin` and is available in TinyGo builds.

See the [ST7036 datasheet](https://www.crystalfontz.com/controllers/Sitronix/ST7036/).
