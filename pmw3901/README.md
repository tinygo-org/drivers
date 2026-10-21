# PMW3901

This package reads motion from the PixArt PMW3901 optical flow sensor over SPI.

Use `New(bus, cs)`, then call `Configure()`. Configure the bus in mode 3 at up
to 2 MHz, and configure `cs` as an output set high. `Configure` resets the
device, checks the product ID, and writes the performance optimization
registers. These values are from the
[Bitcraze PMW3901 driver](https://github.com/bitcraze/Bitcraze_PMW3901) and are
covered by the license in `LICENSE-bitcraze`.

`ReadMotionBurst` returns the motion counts with the surface quality and
shutter values. Use `Motion.Valid()` to skip readings without motion or with a
poor surface. `ReadMotion` reads only the motion counts. The sensor works at a
distance of 80 mm or more.

`ReadFrame` captures a raw 35x35 frame with 8-bit pixels. It is slow, since the
sensor delivers about one pixel every 7 ms, so a frame takes about 9 seconds.
It stops motion tracking, so call `Configure` again before reading motion. The
capture sequence is from the
[Pimoroni pmw3901-python driver](https://github.com/pimoroni/pmw3901-python) and
is covered by the license in `LICENSE-pimoroni`.

See the [PMW3901 datasheet](https://www.codico.com/de/mpattachment/file/download/id/952/)
for the register map and SPI timings.
