# SN3218 LED driver

This package supports the SI-EN SN3218 18-channel PWM LED driver.
The Pimoroni Display-o-Tron HAT uses this chip for its RGB backlight.

Connect the chip to an I2C bus. The fixed 7-bit address is `0x54`.
Pass the bus to `New`, then call `Configure`. Configuration resets the
registers, enables all 18 channels, and enables the output stage.

`SetChannels` accepts an array of 18 PWM values in channel order.
Each value is from 0 to 255. The driver applies the values with an
update register write. It does not apply gamma correction.

`EnableChannels` uses bits 0 to 17 to select the enabled channels.
Bits above 17 are ignored. The driver applies this mask with an update
register write. `Disable` turns off the output stage. `Enable` turns it
on again. `Reset` restores the registers to their power-on state.

Each operation returns the first I2C error and stops further writes.
A failed update write can leave new values pending in the chip.
The driver does not retry failed writes.

References:

- [SN3218 datasheet](https://www.si-en.com/uploadpdf/s2011528172924.pdf)
- [Pimoroni SN3218 driver](https://github.com/pimoroni/sn3218)
