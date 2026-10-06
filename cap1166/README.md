# CAP1166

This package controls the six capacitive touch inputs and six LED outputs of
the Microchip CAP1166. It uses the register settings from the
[Pimoroni Python driver](https://github.com/pimoroni/cap1xxx-python).

Use `New(bus, address)` with a `drivers.I2C` bus, then call `Configure()`.
`DefaultAddress` is `0x2C`, as used on the Display-o-Tron HAT. Select the address
for your board. Other boards can use `0x28`.

`InputStatus()` returns a six-bit touch mask. `Pressed(channel)` checks one
channel. The `Cancel`, `Up`, `Down`, `Left`, `Button`, and `Right` constants are
channel numbers for the Display-o-Tron HAT, not bit masks. Call
`ClearInterrupt()` after you read the inputs so the chip can update the status.
Repeat events are disabled by default. Use `EnableRepeat(mask)` to enable them.
Use `EnableMultitouch()` and `SetSensitivity()` to change the touch settings.

`SetGraph(percentage)` controls the HAT's LED bar graph. Values below zero are
set to zero. Values above one are set to one. The first LED is output 6. A
partly lit boundary LED uses the direct duty cycle. `GraphOff()` turns off all
six outputs.

Register operations return I2C errors. `Connected()` returns false if the read
fails or the product ID does not match.

See the [CAP1166 datasheet](https://ww1.microchip.com/downloads/en/DeviceDoc/00001621B.pdf)
for the register map and board connections.
