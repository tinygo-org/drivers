# TinyGo drivers

[![GoDoc](https://godoc.org/github.com/aykevl/tinygo-drivers?status.svg)](https://godoc.org/github.com/aykevl/tinygo-drivers)

This package provides a collection of hardware drivers that can be used together
with [TinyGo](https://github.com/aykevl/tinygo).

## Scope

The drivers repository provides self-contained drivers designed for
microcontrollers and bare metal systems, but may also be useful for other
embedded systems such as embedded Linux.

While this repository provides drivers, many of these drivers are not meant to
be used directly. For example:

  * Gyroscope and accelerometer measurements must be
    [fused together](https://en.wikipedia.org/wiki/Sensor_fusion) with something
    like a [complementary](http://www.pieter-jan.com/node/11) or
    [Kalman](https://en.wikipedia.org/wiki/Kalman_filter) filter to be useful.
  * A display should likely be used together with a graphics library that works
    with all provided display drivers.

Such algorithms and libraries are out of scope for the drivers project.

## Contributing

This collection of drivers is part of the
[TinyGo](https://github.com/aykevl/tinygo) project. Patches are welcome but new
drivers must try to follow patterns established by similar existing drivers.
